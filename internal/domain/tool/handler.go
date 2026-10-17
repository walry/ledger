package tool

import (
	"context"
	"encoding/json"
	"errors"
)

// InvocationResult 是工具调用结果（值对象）。
type InvocationResult struct {
	Content  string
	IsError  bool
	Metadata map[string]string
}

// Handler 是工具执行端口，实现在基础设施层（infrastructure/toolhandler）。
type Handler interface {
	// Name 返回处理器名，需与 Tool 的 handler 名一致。
	Name() string
	// Invoke 执行工具，args 为模型给出的 JSON 参数。
	Invoke(ctx context.Context, args json.RawMessage) (InvocationResult, error)
}

// HandlerRef 是对处理器的引用（值对象）。
type HandlerRef struct {
	name string
}

// NewHandlerRef 创建处理器引用。
func NewHandlerRef(name string) (HandlerRef, error) {
	if name == "" {
		return HandlerRef{}, errors.New("tool: handler name is required")
	}
	return HandlerRef{name: name}, nil
}

// Name 返回处理器名。
func (h HandlerRef) Name() string { return h.name }

// IsZero 判断引用是否为空。
func (h HandlerRef) IsZero() bool { return h.name == "" }

// HandlerFactory 是处理器工厂端口，按名创建 Handler。
type HandlerFactory interface {
	Create(name string) (Handler, error)
	Has(name string) bool
	Names() []string
}
