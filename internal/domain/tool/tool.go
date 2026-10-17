package tool

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/walry/ledger/internal/domain/ddd"
)

// Tool 是工具聚合根。
//
// 不变量：name、schema、handler 必填；名称唯一。
type Tool struct {
	id          ToolId
	name        string
	description string
	schema      JSONSchema
	source      ToolSource
	handler     HandlerRef
	enabled     bool
	version     int
	timeout     time.Duration
	createdAt   time.Time
	updatedAt   time.Time
	events      []ddd.DomainEvent
}

// NewTool 创建工具。
func NewTool(id ToolId, name, description string, schema JSONSchema, source ToolSource, handler HandlerRef, timeout time.Duration, now time.Time) (*Tool, error) {
	if id.IsZero() {
		return nil, errors.New("tool: id is required")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("tool: name is required")
	}
	if !schema.Valid() {
		return nil, errors.New("tool: schema is required")
	}
	if handler.IsZero() {
		return nil, errors.New("tool: handler is required")
	}
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	t := &Tool{
		id:          id,
		name:        name,
		description: description,
		schema:      schema,
		source:      source,
		handler:     handler,
		enabled:     true,
		version:     1,
		timeout:     timeout,
		createdAt:   now,
		updatedAt:   now,
	}
	t.record(ToolRegistered{ToolID: id, Name: name, Occurred: now})
	return t, nil
}

// RehydrateTool 从持久化状态重建工具。
func RehydrateTool(
	id ToolId, name, description string, schema JSONSchema, source ToolSource,
	handler HandlerRef, enabled bool, version int, timeout time.Duration,
	createdAt, updatedAt time.Time,
) *Tool {
	return &Tool{
		id: id, name: name, description: description, schema: schema, source: source,
		handler: handler, enabled: enabled, version: version, timeout: timeout,
		createdAt: createdAt, updatedAt: updatedAt,
	}
}

// ID 返回聚合标识。
func (t *Tool) ID() ToolId { return t.id }

// Name 返回工具名。
func (t *Tool) Name() string { return t.name }

// Description 返回描述。
func (t *Tool) Description() string { return t.description }

// Schema 返回参数 Schema。
func (t *Tool) Schema() JSONSchema { return t.schema }

// Source 返回来源。
func (t *Tool) Source() ToolSource { return t.source }

// Handler 返回处理器引用。
func (t *Tool) Handler() HandlerRef { return t.handler }

// Enabled 返回是否启用。
func (t *Tool) Enabled() bool { return t.enabled }

// Version 返回版本。
func (t *Tool) Version() int { return t.version }

// Timeout 返回超时。
func (t *Tool) Timeout() time.Duration { return t.timeout }

// CreatedAt 返回创建时间。
func (t *Tool) CreatedAt() time.Time { return t.createdAt }

// UpdatedAt 返回更新时间。
func (t *Tool) UpdatedAt() time.Time { return t.updatedAt }

// UpdateDefinition 更新描述与 Schema，并递增版本。
func (t *Tool) UpdateDefinition(description string, schema JSONSchema, now time.Time) error {
	if !schema.Valid() {
		return fmt.Errorf("tool: schema is required")
	}
	t.description = description
	t.schema = schema
	t.version++
	t.updatedAt = now
	return nil
}

// Enable 启用工具。
func (t *Tool) Enable(now time.Time) {
	if t.enabled {
		return
	}
	t.enabled = true
	t.updatedAt = now
	t.record(ToolEnabledChanged{ToolID: t.id, Enabled: true, Occurred: now})
}

// Disable 停用工具。
func (t *Tool) Disable(now time.Time) {
	if !t.enabled {
		return
	}
	t.enabled = false
	t.updatedAt = now
	t.record(ToolEnabledChanged{ToolID: t.id, Enabled: false, Occurred: now})
}

// PullEvents 返回并清空领域事件。
func (t *Tool) PullEvents() []ddd.DomainEvent {
	events := t.events
	t.events = nil
	return events
}

func (t *Tool) record(e ddd.DomainEvent) { t.events = append(t.events, e) }
