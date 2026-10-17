package toolhandler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/walry/ledger/internal/domain/tool"
)

const (
	webSearchMaxBodyBytes = 8 * 1024
	webSearchTimeout      = 15 * time.Second
)

var _ tool.Handler = (*WebSearchHandler)(nil)

// WebSearchHandler 通过可配置的搜索端点执行查询。
//
// 端点未配置时，Invoke 返回明确的“未配置”提示而不是报错，
// 以便 Agent 感知能力缺失后改用其他手段。
type WebSearchHandler struct {
	endpoint string
	apiKey   string
	client   *http.Client
}

// NewWebSearchHandler 创建网络搜索处理器。
func NewWebSearchHandler(endpoint, apiKey string) tool.Handler {
	return &WebSearchHandler{
		endpoint: strings.TrimSpace(endpoint),
		apiKey:   strings.TrimSpace(apiKey),
		client:   &http.Client{Timeout: webSearchTimeout},
	}
}

// Name 返回处理器名。
func (h *WebSearchHandler) Name() string { return "web_search" }

// Invoke 执行搜索查询。
func (h *WebSearchHandler) Invoke(ctx context.Context, args json.RawMessage) (tool.InvocationResult, error) {
	if err := ctx.Err(); err != nil {
		return tool.InvocationResult{}, err
	}
	if h.endpoint == "" {
		return errorResult("web search tool is not configured"), nil
	}
	var in struct {
		Query string `json:"query"`
	}
	if err := json.Unmarshal(args, &in); err != nil {
		return errorResult("invalid arguments: " + err.Error()), nil
	}
	query := strings.TrimSpace(in.Query)
	if query == "" {
		return errorResult("query is required"), nil
	}

	u, err := url.Parse(h.endpoint)
	if err != nil {
		return errorResult("invalid search endpoint: " + err.Error()), nil
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return errorResult("unsupported search endpoint scheme"), nil
	}
	q := u.Query()
	q.Set("q", query)
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return errorResult("failed to build request: " + err.Error()), nil
	}
	if h.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+h.apiKey)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := h.client.Do(req)
	if err != nil {
		return errorResult("search request failed: " + err.Error()), nil
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, webSearchMaxBodyBytes+1))
	if err != nil {
		return errorResult("failed to read search response: " + err.Error()), nil
	}
	truncated := len(body) > webSearchMaxBodyBytes
	if truncated {
		body = body[:webSearchMaxBodyBytes]
	}
	content := string(body)
	if truncated {
		content += "\n... [truncated]"
	}
	return tool.InvocationResult{
		Content: content,
		IsError: resp.StatusCode >= http.StatusBadRequest,
		Metadata: map[string]string{
			"status": fmt.Sprintf("%d", resp.StatusCode),
		},
	}, nil
}
