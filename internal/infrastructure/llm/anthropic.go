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

// anthropicVersion 是 Anthropic Messages API 的固定版本头。
const anthropicVersion = "2023-06-01"

// anthropicDefaultMaxTokens 是 max_tokens 缺省值（Anthropic 必填）。
const anthropicDefaultMaxTokens = 1024

// AnthropicClient 通过 Anthropic Messages 协议调用模型。
type AnthropicClient struct {
	baseURL string
	apiKey  string
	model   string
	http    *http.Client
}

// NewAnthropicClient 构造 Anthropic 客户端。baseURL 末尾斜杠会被裁剪。
func NewAnthropicClient(baseURL, apiKey, model string) *AnthropicClient {
	return &AnthropicClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		apiKey:  apiKey,
		model:   model,
		http:    &http.Client{},
	}
}

var _ modelconfig.ClientPort = (*AnthropicClient)(nil)

// ---------------- wire types ----------------

type anthropicMessage struct {
	Role    string `json:"role"`
	Content any    `json:"content"`
}

type anthropicTool struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	InputSchema json.RawMessage `json:"input_schema"`
}

type anthropicTextBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type anthropicToolUseBlock struct {
	Type  string          `json:"type"`
	ID    string          `json:"id"`
	Name  string          `json:"name"`
	Input json.RawMessage `json:"input"`
}

type anthropicToolResultBlock struct {
	Type      string `json:"type"`
	ToolUseID string `json:"tool_use_id"`
	Content   string `json:"content"`
}

type anthropicRequest struct {
	Model         string             `json:"model"`
	System        string             `json:"system,omitempty"`
	Messages      []anthropicMessage `json:"messages"`
	MaxTokens     int                `json:"max_tokens"`
	Temperature   *float64           `json:"temperature,omitempty"`
	TopP          *float64           `json:"top_p,omitempty"`
	StopSequences []string           `json:"stop_sequences,omitempty"`
	Tools         []anthropicTool    `json:"tools,omitempty"`
	Stream        bool               `json:"stream,omitempty"`
}

type anthropicError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

type anthropicResponse struct {
	Content []struct {
		Type  string          `json:"type"`
		Text  string          `json:"text"`
		ID    string          `json:"id"`
		Name  string          `json:"name"`
		Input json.RawMessage `json:"input"`
	} `json:"content"`
	StopReason string `json:"stop_reason"`
	Usage      struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
	Error *anthropicError `json:"error"`
}

type anthropicStreamEvent struct {
	Type  string `json:"type"`
	Index int    `json:"index"`
	Delta struct {
		Type        string `json:"type"`
		Text        string `json:"text"`
		PartialJSON string `json:"partial_json"`
		StopReason  string `json:"stop_reason"`
	} `json:"delta"`
	ContentBlock struct {
		Type string `json:"type"`
		Text string `json:"text"`
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"content_block"`
	Message struct {
		Usage struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	} `json:"message"`
	Usage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
	Error *anthropicError `json:"error"`
}

// ---------------- mapping ----------------

// toAnthropicMessages 将领域消息映射为 Anthropic 消息序列。
// system 角色消息被并入顶层 system，工具结果折叠为 user/tool_result 块。
func toAnthropicMessages(req modelconfig.Request) (string, []anthropicMessage) {
	systemParts := make([]string, 0, 2)
	if strings.TrimSpace(req.System) != "" {
		systemParts = append(systemParts, req.System)
	}

	msgs := make([]anthropicMessage, 0, len(req.Messages))
	for _, m := range req.Messages {
		switch m.Role {
		case modelconfig.RoleSystem:
			if strings.TrimSpace(m.Content) != "" {
				systemParts = append(systemParts, m.Content)
			}
		case modelconfig.RoleTool:
			msgs = append(msgs, anthropicMessage{
				Role: string(modelconfig.RoleUser),
				Content: []anthropicToolResultBlock{{
					Type:      "tool_result",
					ToolUseID: m.ToolCallID,
					Content:   m.Content,
				}},
			})
		case modelconfig.RoleAssistant:
			msgs = append(msgs, anthropicMessage{
				Role:    string(modelconfig.RoleAssistant),
				Content: anthropicAssistantContent(m),
			})
		default:
			msgs = append(msgs, anthropicMessage{
				Role:    string(m.Role),
				Content: m.Content,
			})
		}
	}
	return strings.Join(systemParts, "\n\n"), msgs
}

// anthropicAssistantContent 返回 assistant 消息的 content：纯文本用字符串，
// 含工具调用时用 text/tool_use 块数组。
func anthropicAssistantContent(m modelconfig.Message) any {
	if len(m.ToolCalls) == 0 {
		return m.Content
	}
	blocks := make([]any, 0, len(m.ToolCalls)+1)
	if strings.TrimSpace(m.Content) != "" {
		blocks = append(blocks, anthropicTextBlock{Type: "text", Text: m.Content})
	}
	for _, tc := range m.ToolCalls {
		blocks = append(blocks, anthropicToolUseBlock{
			Type:  "tool_use",
			ID:    tc.ID,
			Name:  tc.Name,
			Input: rawJSONOrEmptyObject(tc.Arguments),
		})
	}
	return blocks
}

// rawJSONOrEmptyObject 保证 Arguments 是合法 JSON 对象，否则退化为 {}。
func rawJSONOrEmptyObject(arguments string) json.RawMessage {
	trimmed := strings.TrimSpace(arguments)
	if trimmed == "" || !json.Valid([]byte(trimmed)) {
		return json.RawMessage(`{}`)
	}
	return json.RawMessage(trimmed)
}

func toAnthropicTools(tools []modelconfig.ToolSchema) []anthropicTool {
	out := make([]anthropicTool, 0, len(tools))
	for _, t := range tools {
		schema := t.Parameters
		if len(schema) == 0 {
			schema = defaultToolParameters
		}
		out = append(out, anthropicTool{
			Name:        t.Name,
			Description: t.Description,
			InputSchema: schema,
		})
	}
	return out
}

// ---------------- ClientPort ----------------

// Invoke 执行一次非流式调用。
func (c *AnthropicClient) Invoke(ctx context.Context, req modelconfig.Request) (modelconfig.Response, error) {
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
		return modelconfig.Response{}, fmt.Errorf("llm/anthropic: request failed: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return modelconfig.Response{}, fmt.Errorf("llm/anthropic: read response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return modelconfig.Response{}, &apiError{Provider: "anthropic", Status: resp.StatusCode, Message: decodeErrorMessage(data)}
	}

	var parsed anthropicResponse
	if err := json.Unmarshal(data, &parsed); err != nil {
		return modelconfig.Response{}, fmt.Errorf("llm/anthropic: decode response: %w", err)
	}
	if parsed.Error != nil {
		return modelconfig.Response{}, &apiError{Provider: "anthropic", Message: parsed.Error.Message}
	}

	out := modelconfig.Response{
		FinishReason: parsed.StopReason,
		Usage: modelconfig.Usage{
			PromptTokens:     parsed.Usage.InputTokens,
			CompletionTokens: parsed.Usage.OutputTokens,
			TotalTokens:      parsed.Usage.InputTokens + parsed.Usage.OutputTokens,
		},
	}
	var text strings.Builder
	for _, block := range parsed.Content {
		switch block.Type {
		case "text":
			text.WriteString(block.Text)
		case "tool_use":
			out.ToolCalls = append(out.ToolCalls, modelconfig.ToolCall{
				ID:        block.ID,
				Name:      block.Name,
				Arguments: string(block.Input),
			})
		}
	}
	out.Content = text.String()
	return out, nil
}

// Stream 执行流式调用，返回带缓冲的增量通道。
func (c *AnthropicClient) Stream(ctx context.Context, req modelconfig.Request) (<-chan modelconfig.Chunk, error) {
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
		return nil, fmt.Errorf("llm/anthropic: request failed: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer resp.Body.Close()
		data, _ := io.ReadAll(resp.Body)
		return nil, &apiError{Provider: "anthropic", Status: resp.StatusCode, Message: decodeErrorMessage(data)}
	}

	ch := make(chan modelconfig.Chunk, streamBuffer)
	go func() {
		defer close(ch)
		defer resp.Body.Close()

		acc := newToolCallAccumulator()
		var promptTokens, completionTokens int

		scanErr := scanSSE(ctx, resp.Body, func(data []byte) error {
			var ev anthropicStreamEvent
			if err := json.Unmarshal(data, &ev); err != nil {
				return fmt.Errorf("llm/anthropic: decode stream event: %w", err)
			}
			switch ev.Type {
			case "message_start":
				promptTokens = ev.Message.Usage.InputTokens
				completionTokens = ev.Message.Usage.OutputTokens
			case "content_block_start":
				if ev.ContentBlock.Type == "tool_use" {
					acc.add(ev.Index, ev.ContentBlock.ID, ev.ContentBlock.Name, "")
				} else if ev.ContentBlock.Text != "" {
					if !sendChunk(ctx, ch, modelconfig.Chunk{Delta: ev.ContentBlock.Text}) {
						return ctx.Err()
					}
				}
			case "content_block_delta":
				switch ev.Delta.Type {
				case "text_delta":
					if ev.Delta.Text != "" {
						if !sendChunk(ctx, ch, modelconfig.Chunk{Delta: ev.Delta.Text}) {
							return ctx.Err()
						}
					}
				case "input_json_delta":
					acc.add(ev.Index, "", "", ev.Delta.PartialJSON)
				}
			case "message_delta":
				if ev.Usage.OutputTokens > 0 {
					completionTokens = ev.Usage.OutputTokens
				}
				if ev.Usage.InputTokens > 0 {
					promptTokens = ev.Usage.InputTokens
				}
			case "error":
				if ev.Error != nil {
					return &apiError{Provider: "anthropic", Message: ev.Error.Message}
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
		usage := &modelconfig.Usage{
			PromptTokens:     promptTokens,
			CompletionTokens: completionTokens,
			TotalTokens:      promptTokens + completionTokens,
		}
		sendChunk(ctx, ch, modelconfig.Chunk{Done: true, Usage: usage})
	}()
	return ch, nil
}

// ---------------- helpers ----------------

func (c *AnthropicClient) encodeRequest(req modelconfig.Request, stream bool) ([]byte, error) {
	system, messages := toAnthropicMessages(req)

	maxTokens := req.Params.MaxTokens
	if maxTokens <= 0 {
		maxTokens = anthropicDefaultMaxTokens
	}
	temperature := req.Params.Temperature
	topP := req.Params.TopP

	payload := anthropicRequest{
		Model:         c.resolveModel(req.Model),
		System:        system,
		Messages:      messages,
		MaxTokens:     maxTokens,
		Temperature:   &temperature,
		TopP:          &topP,
		StopSequences: req.Params.Stop,
		Tools:         toAnthropicTools(req.Tools),
		Stream:        stream,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("llm/anthropic: encode request: %w", err)
	}
	return body, nil
}

func (c *AnthropicClient) newHTTPRequest(ctx context.Context, body []byte) (*http.Request, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/messages", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("llm/anthropic: build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("anthropic-version", anthropicVersion)
	return httpReq, nil
}

func (c *AnthropicClient) resolveModel(reqModel string) string {
	if strings.TrimSpace(reqModel) != "" {
		return reqModel
	}
	return c.model
}
