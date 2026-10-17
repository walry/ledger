package agent

import (
	"time"

	"github.com/walry/ledger/internal/domain/ddd"
)

// AgentCreated 智能体已创建。
type AgentCreated struct {
	AgentID  AgentId
	Name     string
	Version  int
	Occurred time.Time
}

// EventName 实现 ddd.DomainEvent。
func (AgentCreated) EventName() string { return "AgentCreated" }

// OccurredAt 实现 ddd.DomainEvent。
func (e AgentCreated) OccurredAt() time.Time { return e.Occurred }

// AgentPublished 智能体已发布，缓存可据此刷新。
type AgentPublished struct {
	AgentID  AgentId
	Version  int
	Occurred time.Time
}

// EventName 实现 ddd.DomainEvent。
func (AgentPublished) EventName() string { return "AgentPublished" }

// OccurredAt 实现 ddd.DomainEvent。
func (e AgentPublished) OccurredAt() time.Time { return e.Occurred }

// AgentArchived 智能体已归档。
type AgentArchived struct {
	AgentID  AgentId
	Version  int
	Occurred time.Time
}

// EventName 实现 ddd.DomainEvent。
func (AgentArchived) EventName() string { return "AgentArchived" }

// OccurredAt 实现 ddd.DomainEvent。
func (e AgentArchived) OccurredAt() time.Time { return e.Occurred }

var _ ddd.DomainEvent = AgentCreated{}
var _ ddd.DomainEvent = AgentPublished{}
var _ ddd.DomainEvent = AgentArchived{}
