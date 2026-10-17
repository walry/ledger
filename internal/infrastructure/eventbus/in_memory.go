// Package eventbus 提供 ddd.EventBus 端口的内存实现。
//
// 订阅关系只在装配层注册；本实现并发安全，Handler 失败时记录并继续，
// 最终以 errors.Join 聚合返回，符合“事件可能重复投递、Handler 必须幂等”的约定。
package eventbus

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/walry/ledger/internal/domain/ddd"
)

var _ ddd.EventBus = (*InMemoryBus)(nil)

// InMemoryBus 是进程内的同步事件总线。
type InMemoryBus struct {
	mu       sync.RWMutex
	handlers map[string][]ddd.Handler
}

// NewInMemoryBus 创建内存事件总线。
func NewInMemoryBus() *InMemoryBus {
	return &InMemoryBus{handlers: make(map[string][]ddd.Handler)}
}

// Subscribe 为一个事件名注册处理器。并发安全。
func (b *InMemoryBus) Subscribe(eventName string, h ddd.Handler) {
	if h == nil {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[eventName] = append(b.handlers[eventName], h)
}

// Publish 按 EventName 将每个事件分发给其订阅者。
//
// 处理器可能失败：失败会被记录并继续执行其余处理器，最终返回聚合错误。
func (b *InMemoryBus) Publish(ctx context.Context, events ...ddd.DomainEvent) error {
	var errs []error
	for _, event := range events {
		if event == nil {
			continue
		}
		b.mu.RLock()
		handlers := make([]ddd.Handler, len(b.handlers[event.EventName()]))
		copy(handlers, b.handlers[event.EventName()])
		b.mu.RUnlock()

		for _, h := range handlers {
			if h == nil {
				continue
			}
			if err := h(ctx, event); err != nil {
				log.Printf("eventbus: handler for %s failed: %v", event.EventName(), err)
				errs = append(errs, fmt.Errorf("eventbus: handler for %s: %w", event.EventName(), err))
			}
		}
	}
	return errors.Join(errs...)
}
