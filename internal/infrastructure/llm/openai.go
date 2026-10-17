package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/walry/ledger/internal/domain/modelconfig"
)

// OpenAIClient 通过 OpenAI Chat Completions 协议调用模型。
// 同时适用于任何兼容 OpenAI 协议的第三方服务。
type OpenAIClient struct {
	baseURL string
	apiKey  string
	model   string
	http    *http.Client
}

// NewOpenAIClient 构造 OpenAI 客户端。baseURL 末尾斜杠会被裁剪。
func NewOpenAIClient(baseURL, apiKey, model string) *OpenAIClient {
	return &OpenAIClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		apiKey:  apiKey,
		model:   model,
		http:    &http.Client{},
	}
}

var _ modelconfig.ClientPort = (*OpenAIClient)(nil)

// ---------------- wire types ----------------

type openAIMessage struct {
	Role       string           `json:"role"`
	Content    string           `json:"content,omitempty"`
	Name       string           `json:"name,omitempty"`
	ToolCalls  []openAIToolCall `json:"tool_calls,omitempty"`
	ToolCallID string           `json:"tool_call_id,omitempty"`
}

type openAIToolCall struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"`
	Function openAIToolCallFunction `json:"function"`
}

type openAIToolCallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type openAITool struct {
	Type     string             `json:"type"`
	Function openAIToolFunction `json:"function"`
}

type openAIToolFunction struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Parameters  json.RawMessage `json:"parameters,omitempty"`
}

type openAIStreamOptions struct {
	IncludeUsage bool `json:"include_usage"`
}

type openAIChatRequest struct {
	Model            string               `json:"model"`
	Messages         []openAIMessage      `json:"messages"`
	Temperature      float64              `json:"temperature"`
	TopP             float64              `json:"top_p"`
	MaxTokens        int                  `json:"max_tokens,omitempty"`
	PresencePenalty  float64              `json:"presence_penalty"`
	FrequencyPenalty float64              `json:"frequency_penalty"`
	Stop             []string             `json:"stop,omitempty"`
	Tools            []openAITool         `json:"tools,omitempty"`
	Stream           bool                 `json:"stream,omitempty"`
	StreamOptions    *openAIStreamOptions `json:"stream_options,omitempty"`
}

type openAIError struct {
	Message string `json:"message"`
}

type openAIChatResponse struct {
	Choices []struct {
		Message struct {
			Content   string           `json:"content"`
			ToolCalls []openAIToolCall `json:"tool_calls"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Error *openAIError `json:"error"`
}

type openAIStreamChunk struct {
	Choices []struct {
		Delta struct {
			Content   string `json:"content"`
			ToolCalls []struct {
				Index    int    `json:"index"`
				ID       string `json:"id"`
				Type     string `json:"type"`
				Function struct {
					Name      string `json:"name"`
					Arguments string `json:"arguments"`
				} `json:"function"`
			} `json:"tool_calls"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage *struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Error *openAIError `json:"error"`
}

// defaultToolParameters 用于工具未声明参数 schema 时提供合法空对象。
var defaultToolParameters = json.RawMessage(`{"type":"object","properties":{}}`)

// ---------------- mapping ----------------

func toOpenAIMessages(req modelconfig.Request) []openAIMessage {
	msgs := make([]openAIMessage, 0, len(req.Messages)+1)
	if strings.TrimSpace(req.System) != "" {
		msgs = append(msgs, openAIMessage{Role: string(modelconfig.RoleSystem), Content: req.System})
	}
	for _, m := range req.Messages {
		om := openAIMessage{
			Role:       string(m.Role),
			Content:    m.Content,
			Name:       m.Name,
			ToolCallID: m.ToolCallID,
		}
		for _, tc := range m.ToolCalls {
			om.ToolCalls = append(om.ToolCalls, openAIToolCall{
				ID:       tc.ID,
				Type:     "function",
				Function: openAIToolCallFunction{Name: tc.Name, Arguments: tc.Arguments},
			})
		}
		msgs = append(msgs, om)
	}
	return msgs
}

func toOpenAITools(tools []modelconfig.ToolSchema) []openAITool {
	out := make([]openAITool, 0, len(tools))
	for _, t := range tools {
		params := t.Parameters
		if len(params) == 0 {
			params = defaultToolParameters
		}
		out = append(out, openAITool{
			Type: "function",
			Function: openAIToolFunction{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  params,
			},
		})
	}
	return out
}

// ---------------- ClientPort ----------------

// Invoke 执行一次非流式调用。
func (c *OpenAIClient) Invoke(ctx context.Context, req modelconfig.Request) (modelconfig.Response, error) {
	body, err := c.encodeRequest(req, false)
	if err != nil {
		return modelconfig.Response{}, err
	}
	httpReq, err := c.newHTTPRequest(ctx, body)
	if err != nil {
		return modelconfig.Response{}, err
	}
	resp, err := c.http.Do(httpReq)
	if err != nil {
		return modelconfig.Response{}, fmt.Errorf("llm/openai: request failed: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return modelconfig.Response{}, fmt.Errorf("llm/openai: read response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return modelconfig.Response{}, &apiError{Provider: "openai", Status: resp.StatusCode, Message: decodeErrorMessage(data)}
	}

	var parsed openAIChatResponse
	if err := json.Unmarshal(data, &parsed); err != nil {
		return modelconfig.Response{}, fmt.Errorf("llm/openai: decode response: %w", err)
	}
	if parsed.Error != nil {
		return modelconfig.Response{}, &apiError{Provider: "openai", Message: parsed.Error.Message}
	}
	if len(parsed.Choices) == 0 {
		return modelconfig.Response{}, &apiError{Provider: "openai", Message: "empty choices in response"}
	}

	choice := parsed.Choices[0]
	out := modelconfig.Response{
		Content:      choice.Message.Content,
		FinishReason: choice.FinishReason,
		Usage: modelconfig.Usage{
			PromptTokens:     parsed.Usage.PromptTokens,
			CompletionTokens: parsed.Usage.CompletionTokens,
			TotalTokens:      parsed.Usage.TotalTokens,
		},
	}
	for _, tc := range choice.Message.ToolCalls {
		out.ToolCalls = append(out.ToolCalls, modelconfig.ToolCall{
			ID:        tc.ID,
			Name:      tc.Function.Name,
			Arguments: tc.Function.Arguments,
		})
	}
	return out, nil
}

// Stream 执行流式调用，返回带缓冲的增量通道。
func (c *OpenAIClient) Stream(ctx context.Context, req modelconfig.Request) (<-chan modelconfig.Chunk, error) {
	body, err := c.encodeRequest(req, true)
	if err != nil {
		return nil, err
	}
	httpReq, err := c.newHTTPRequest(ctx, body)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Accept", "text/event-stream")

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("llm/openai: request failed: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer resp.Body.Close()
		data, _ := io.ReadAll(resp.Body)
		return nil, &apiError{Provider: "openai", Status: resp.StatusCode, Message: decodeErrorMessage(data)}
	}

	ch := make(chan modelconfig.Chunk, streamBuffer)
	go func() {
		defer close(ch)
		defer resp.Body.Close()

		acc := newToolCallAccumulator()
		var usage *modelconfig.Usage

		scanErr := scanSSE(ctx, resp.Body, func(data []byte) error {
			var ev openAIStreamChunk
			if err := json.Unmarshal(data, &ev); err != nil {
				return fmt.Errorf("llm/openai: decode stream chunk: %w", err)
			}
			if ev.Error != nil {
				return &apiError{Provider: "openai", Message: ev.Error.Message}
			}
			if len(ev.Choices) > 0 {
				delta := ev.Choices[0].Delta
				if delta.Content != "" {
					if !sendChunk(ctx, ch, modelconfig.Chunk{Delta: delta.Content}) {
						return ctx.Err()
					}
				}
				for _, tc := range delta.ToolCalls {
					acc.add(tc.Index, tc.ID, tc.Function.Name, tc.Function.Arguments)
				}
			}
			if ev.Usage != nil {
				usage = &modelconfig.Usage{
					PromptTokens:     ev.Usage.PromptTokens,
					CompletionTokens: ev.Usage.CompletionTokens,
					TotalTokens:      ev.Usage.TotalTokens,
				}
			}
			return nil
		})
		if scanErr != nil {
			sendChunk(ctx, ch, modelconfig.Chunk{Err: scanErr})
			return
		}
		if !acc.empty() && !emitToolCalls(ctx, ch, acc.result()) {
			return
		}
		sendChunk(ctx, ch, modelconfig.Chunk{Done: true, Usage: usage})
	}()
	return ch, nil
}

// ---------------- helpers ----------------

func (c *OpenAIClient) encodeRequest(req modelconfig.Request, stream bool) ([]byte, error) {
	payload := openAIChatRequest{
		Model:            c.resolveModel(req.Model),
		Messages:         toOpenAIMessages(req),
		Temperature:      req.Params.Temperature,
		TopP:             req.Params.TopP,
		MaxTokens:        req.Params.MaxTokens,
		PresencePenalty:  req.Params.PresencePenalty,
		FrequencyPenalty: req.Params.FrequencyPenalty,
		Stop:             req.Params.Stop,
		Tools:            toOpenAITools(req.Tools),
		Stream:           stream,
	}
	if stream {
		payload.StreamOptions = &openAIStreamOptions{IncludeUsage: true}
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("llm/openai: encode request: %w", err)
	}
	return body, nil
}

func (c *OpenAIClient) newHTTPRequest(ctx context.Context, body []byte) (*http.Request, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("llm/openai: build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	return httpReq, nil
}

func (c *OpenAIClient) resolveModel(reqModel string) string {
	if strings.TrimSpace(reqModel) != "" {
		return reqModel
	}
	return c.model
}
