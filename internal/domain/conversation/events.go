package conversation

import (
	"time"

	"github.com/walry/ledger/internal/domain/agent"
	"github.com/walry/ledger/internal/domain/ddd"
)

// ConversationStarted 会话已创建。
type ConversationStarted struct {
	ConversationID ConversationId
	AgentID        agent.AgentId
	AgentVersion   int
	Occurred       time.Time
}

// EventName 实现 ddd.DomainEvent。
func (ConversationStarted) EventName() string { return "ConversationStarted" }

// OccurredAt 实现 ddd.DomainEvent。
func (e ConversationStarted) OccurredAt() time.Time { return e.Occurred }

// TurnFinished 一轮对话完成。
type TurnFinished struct {
	ConversationID ConversationId
	TurnID         TurnId
	Occurred       time.Time
}

// EventName 实现 ddd.DomainEvent。
func (TurnFinished) EventName() string { return "TurnCompleted" }

// OccurredAt 实现 ddd.DomainEvent。
func (e TurnFinished) OccurredAt() time.Time { return e.Occurred }

// ContextCompressed 上下文已压缩。
type ContextCompressed struct {
	ConversationID ConversationId
	UpToTurnIndex  int
	Occurred       time.Time
}

// EventName 实现 ddd.DomainEvent。
func (ContextCompressed) EventName() string { return "ContextCompressed" }

// OccurredAt 实现 ddd.DomainEvent。
func (e ContextCompressed) OccurredAt() time.Time { return e.Occurred }

var _ ddd.DomainEvent = ConversationStarted{}
var _ ddd.DomainEvent = TurnFinished{}
var _ ddd.DomainEvent = ContextCompressed{}
