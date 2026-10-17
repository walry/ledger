package conversation

import (
	"time"

	"github.com/walry/ledger/internal/domain/agent"
)

// TurnStatus 是轮次状态。
type TurnStatus string

const (
	// TurnRunning 进行中。
	TurnRunning TurnStatus = "running"
	// TurnCompleted 已完成。
	TurnCompleted TurnStatus = "completed"
	// TurnFailed 失败。
	TurnFailed TurnStatus = "failed"
)

// Turn 是会话聚合内的实体。每一轮记忆其使用的 Agent 版本（轮级绑定），
// 从而在 Agent 发布新版本后仍可追溯历史对话。
type Turn struct {
	id           TurnId
	index        int
	agentID      agent.AgentId
	agentVersion int
	messages     []Message
	status       TurnStatus
	failure      string
	createdAt    time.Time
	completedAt  *time.Time
}

// NewTurn 创建轮次。
func NewTurn(id TurnId, index int, agentID agent.AgentId, agentVersion int, now time.Time) *Turn {
	return &Turn{
		id:           id,
		index:        index,
		agentID:      agentID,
		agentVersion: agentVersion,
		messages:     make([]Message, 0),
		status:       TurnRunning,
		createdAt:    now,
	}
}

// RehydrateTurn 从持久化状态重建轮次。
func RehydrateTurn(
	id TurnId, index int, agentID agent.AgentId, agentVersion int,
	messages []Message, status TurnStatus, failure string, createdAt time.Time, completedAt *time.Time,
) *Turn {
	if messages == nil {
		messages = make([]Message, 0)
	}
	return &Turn{
		id: id, index: index, agentID: agentID, agentVersion: agentVersion,
		messages: messages, status: status, failure: failure, createdAt: createdAt, completedAt: completedAt,
	}
}

// ID 返回轮次标识。
func (t *Turn) ID() TurnId { return t.id }

// Index 返回轮次序号。
func (t *Turn) Index() int { return t.index }

// AgentID 返回绑定的智能体。
func (t *Turn) AgentID() agent.AgentId { return t.agentID }

// AgentVersion 返回绑定的智能体版本。
func (t *Turn) AgentVersion() int { return t.agentVersion }

// Messages 返回消息列表。
func (t *Turn) Messages() []Message { return t.messages }

// Status 返回状态。
func (t *Turn) Status() TurnStatus { return t.status }

// Failure 返回失败原因。
func (t *Turn) Failure() string { return t.failure }

// CreatedAt 返回创建时间。
func (t *Turn) CreatedAt() time.Time { return t.createdAt }

// CompletedAt 返回完成时间。
func (t *Turn) CompletedAt() *time.Time { return t.completedAt }

// AppendMessage 追加一条消息。
func (t *Turn) AppendMessage(m Message) { t.messages = append(t.messages, m) }

// LastAssistantContent 返回最后一条助手消息内容。
func (t *Turn) LastAssistantContent() string {
	for i := len(t.messages) - 1; i >= 0; i-- {
		if t.messages[i].Role() == RoleAssistant && t.messages[i].Content() != "" {
			return t.messages[i].Content()
		}
	}
	return ""
}

// MarkCompleted 标记完成。
func (t *Turn) MarkCompleted(now time.Time) {
	t.status = TurnCompleted
	t.completedAt = &now
}

// MarkFailed 标记失败。
func (t *Turn) MarkFailed(reason string, now time.Time) {
	t.status = TurnFailed
	t.failure = reason
	t.completedAt = &now
}
