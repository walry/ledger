// Package toolhandler 提供 tool 领域 Handler 端口的内置实现与注册表。
package toolhandler

import (
	"fmt"
	"sort"
	"sync"

	"github.com/walry/ledger/internal/domain/tool"
)

var _ tool.HandlerFactory = (*Registry)(nil)

// Registry 实现 tool.HandlerFactory，按名持有 Handler。并发安全。
type Registry struct {
	mu       sync.RWMutex
	handlers map[string]tool.Handler
}

// NewRegistry 创建空注册表。
func NewRegistry() *Registry {
	return &Registry{handlers: make(map[string]tool.Handler)}
}

// Register 注册处理器，键为其 Name()。
func (r *Registry) Register(h tool.Handler) {
	if h == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.handlers[h.Name()] = h
}

// Create 按名创建（返回已注册的）处理器。
func (r *Registry) Create(name string) (tool.Handler, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	h, ok := r.handlers[name]
	if !ok {
		return nil, fmt.Errorf("toolhandler: handler %q not found", name)
	}
	return h, nil
}

// Has 判断处理器是否存在。
func (r *Registry) Has(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.handlers[name]
	return ok
}

// Names 返回已注册处理器名（有序）。
func (r *Registry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.handlers))
	for name := range r.handlers {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
