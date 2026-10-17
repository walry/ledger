// Package ddd 提供各限界上下文共享的领域事件内核。
//
// 领域层只认识本包定义的 DomainEvent 与 EventBus 接口；
// 具体的事件总线实现位于基础设施层（internal/infrastructure/eventbus），
// 订阅关系只在装配层（cmd/api/main.go）注册。
package ddd

import (
	"context"
	"time"
)

// DomainEvent 是所有领域事件的基接口。
type DomainEvent interface {
	// EventName 返回事件名，用于事件路由（如 "ReportUploaded"）。
	EventName() string
	// OccurredAt 返回事件发生时间。
	OccurredAt() time.Time
}

// EventBus 发布领域事件。
type EventBus interface {
	Publish(ctx context.Context, events ...DomainEvent) error
}

// Handler 处理领域事件。实现必须是幂等的，因为事件可能被重复投递。
type Handler func(ctx context.Context, event DomainEvent) error
