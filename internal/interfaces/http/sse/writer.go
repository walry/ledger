// Package sse 提供一个最小的 Server-Sent Events 写入器。
package sse

import (
	"encoding/json"
	"fmt"
	nethttp "net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// Event 是一条 SSE 事件。
type Event struct {
	Event string
	Data  any
	ID    string
}

// Writer 封装 SSE 响应写入。
type Writer struct {
	w gin.ResponseWriter
}

// NewWriter 初始化 SSE 响应头（禁用缓冲并立即 flush）。
func NewWriter(c *gin.Context) *Writer {
	h := c.Writer.Header()
	h.Set("Content-Type", "text/event-stream")
	h.Set("Cache-Control", "no-cache")
	h.Set("Connection", "keep-alive")
	h.Set("X-Accel-Buffering", "no")
	c.Writer.WriteHeader(nethttp.StatusOK)
	c.Writer.Flush()
	return &Writer{w: c.Writer}
}

// Send 发送一条事件。多行数据会被拆分为多条 data: 行。
func (s *Writer) Send(e Event) error {
	if e.ID != "" {
		if _, err := fmt.Fprintf(s.w, "id: %s\n", e.ID); err != nil {
			return err
		}
	}
	if e.Event != "" {
		if _, err := fmt.Fprintf(s.w, "event: %s\n", e.Event); err != nil {
			return err
		}
	}
	data := ""
	switch v := e.Data.(type) {
	case nil:
		data = ""
	case string:
		data = v
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return err
		}
		data = string(b)
	}
	for _, line := range strings.Split(data, "\n") {
		if _, err := fmt.Fprintf(s.w, "data: %s\n", line); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprint(s.w, "\n"); err != nil {
		return err
	}
	s.w.Flush()
	return nil
}

// Comment 发送注释行，用于心跳保活。
func (s *Writer) Comment(text string) error {
	if _, err := fmt.Fprintf(s.w, ": %s\n\n", text); err != nil {
		return err
	}
	s.w.Flush()
	return nil
}
