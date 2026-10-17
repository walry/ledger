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
	httpMaxBodyBytes = 8 * 1024
	httpTimeout      = 15 * time.Second
)

var _ tool.Handler = (*HTTPHandler)(nil)

// HTTPHandler 执行受限的 HTTP GET/POST 请求。
type HTTPHandler struct {
	client *http.Client
}

// NewHTTPHandler 创建 HTTP 请求处理器。
func NewHTTPHandler() *HTTPHandler {
	return &HTTPHandler{client: &http.Client{Timeout: httpTimeout}}
}

// Name 返回处理器名。
func (h *HTTPHandler) Name() string { return "http_request" }

// Invoke 发起请求，返回截断后的响应体。
func (h *HTTPHandler) Invoke(ctx context.Context, args json.RawMessage) (tool.InvocationResult, error) {
	if err := ctx.Err(); err != nil {
		return tool.InvocationResult{}, err
	}
	var in struct {
		URL     string            `json:"url"`
		Method  string            `json:"method"`
		Headers map[string]string `json:"headers"`
		Body    string            `json:"body"`
	}
	if err := json.Unmarshal(args, &in); err != nil {
		return errorResult("invalid arguments: " + err.Error()), nil
	}
	rawURL := strings.TrimSpace(in.URL)
	if rawURL == "" {
		return errorResult("url is required"), nil
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return errorResult("invalid url: " + err.Error()), nil
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return errorResult("unsupported scheme: only http and https are allowed"), nil
	}

	method := strings.ToUpper(strings.TrimSpace(in.Method))
	if method == "" {
		method = http.MethodGet
	}
	var bodyReader io.Reader
	if in.Body != "" {
		bodyReader = strings.NewReader(in.Body)
	}
	req, err := http.NewRequestWithContext(ctx, method, rawURL, bodyReader)
	if err != nil {
		return errorResult("failed to build request: " + err.Error()), nil
	}
	for k, v := range in.Headers {
		req.Header.Set(k, v)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return errorResult("request failed: " + err.Error()), nil
	}
	defer func() { _ = resp.Body.Close() }()

	limited, err := io.ReadAll(io.LimitReader(resp.Body, httpMaxBodyBytes+1))
	if err != nil {
		return errorResult("failed to read response: " + err.Error()), nil
	}
	truncated := len(limited) > httpMaxBodyBytes
	if truncated {
		limited = limited[:httpMaxBodyBytes]
	}

	content := string(limited)
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
