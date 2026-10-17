package embedding

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/walry/ledger/internal/domain/knowledge"
	"github.com/walry/ledger/internal/domain/knowledge/service"
)

// defaultBaseURL 是 baseURL 缺省时的 OpenAI 服务地址。
const defaultBaseURL = "https://api.openai.com/v1"

// OpenAIEmbedder 通过 OpenAI /embeddings 协议生成文本向量。
type OpenAIEmbedder struct {
	baseURL   string
	apiKey    string
	model     string
	dimension int
	http      *http.Client
}

// NewOpenAIEmbedder 构造 OpenAI 向量化器。baseURL 末尾斜杠会被裁剪，
// 为空时使用 OpenAI 默认地址。
func NewOpenAIEmbedder(baseURL, apiKey, model string, dimension int) *OpenAIEmbedder {
	baseURL = strings.TrimRight(baseURL, "/")
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	return &OpenAIEmbedder{
		baseURL:   baseURL,
		apiKey:    apiKey,
		model:     model,
		dimension: dimension,
		http:      &http.Client{},
	}
}

var _ service.Embedder = (*OpenAIEmbedder)(nil)

type embeddingsRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type embeddingsResponse struct {
	Data []struct {
		Index     int       `json:"index"`
		Embedding []float32 `json:"embedding"`
	} `json:"data"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// Embed 将一批文本转为向量，返回顺序与输入一致。
func (e *OpenAIEmbedder) Embed(ctx context.Context, texts []string) ([]knowledge.Embedding, error) {
	if len(texts) == 0 {
		return []knowledge.Embedding{}, nil
	}

	body, err := json.Marshal(embeddingsRequest{Model: e.model, Input: texts})
	if err != nil {
		return nil, fmt.Errorf("embedding/openai: encode request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, e.baseURL+"/embeddings", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("embedding/openai: build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+e.apiKey)

	resp, err := e.http.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("embedding/openai: request failed: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("embedding/openai: read response: %w", err)
	}

	var parsed embeddingsResponse
	if err := json.Unmarshal(data, &parsed); err != nil {
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return nil, fmt.Errorf("embedding/openai: http %d", resp.StatusCode)
		}
		return nil, fmt.Errorf("embedding/openai: decode response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("embedding/openai: http %d: %s", resp.StatusCode, errorMessage(parsed))
	}
	if parsed.Error != nil {
		return nil, fmt.Errorf("embedding/openai: %s", parsed.Error.Message)
	}
	if len(parsed.Data) != len(texts) {
		return nil, fmt.Errorf("embedding/openai: expected %d embeddings, got %d", len(texts), len(parsed.Data))
	}

	out := make([]knowledge.Embedding, len(texts))
	for _, item := range parsed.Data {
		if item.Index < 0 || item.Index >= len(texts) {
			return nil, fmt.Errorf("embedding/openai: embedding index %d out of range", item.Index)
		}
		emb, err := knowledge.NewEmbedding(e.model, item.Embedding)
		if err != nil {
			return nil, fmt.Errorf("embedding/openai: %w", err)
		}
		out[item.Index] = emb
	}
	return out, nil
}

// Dimension 返回向量维度。
func (e *OpenAIEmbedder) Dimension() int { return e.dimension }

// Model 返回模型名。
func (e *OpenAIEmbedder) Model() string { return e.model }

func errorMessage(resp embeddingsResponse) string {
	if resp.Error != nil && resp.Error.Message != "" {
		return resp.Error.Message
	}
	return "unknown error"
}
