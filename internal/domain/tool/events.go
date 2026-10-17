package tool

import (
	"time"

	"github.com/walry/ledger/internal/domain/ddd"
)

// ToolRegistered 工具已注册。
type ToolRegistered struct {
	ToolID   ToolId
	Name     string
	Occurred time.Time
}

// EventName 实现 ddd.DomainEvent。
func (ToolRegistered) EventName() string { return "ToolRegistered" }

// OccurredAt 实现 ddd.DomainEvent。
func (e ToolRegistered) OccurredAt() time.Time { return e.Occurred }

// ToolEnabledChanged 工具的启用状态发生变化。
type ToolEnabledChanged struct {
	ToolID   ToolId
	Enabled  bool
	Occurred time.Time
}

// EventName 实现 ddd.DomainEvent。
func (ToolEnabledChanged) EventName() string { return "ToolEnabledChanged" }

// OccurredAt 实现 ddd.DomainEvent。
func (e ToolEnabledChanged) OccurredAt() time.Time { return e.Occurred }

var _ ddd.DomainEvent = ToolRegistered{}
var _ ddd.DomainEvent = ToolEnabledChanged{}
