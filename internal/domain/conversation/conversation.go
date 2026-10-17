package conversation

import (
	"errors"
	"fmt"
	"time"

	"github.com/walry/ledger/internal/domain/agent"
	"github.com/walry/ledger/internal/domain/ddd"
	"github.com/walry/ledger/pkg/idgen"
)

// Status 是会话状态。
type Status string

const (
	// StatusActive 活跃。
	StatusActive Status = "active"
	// StatusArchived 已归档。
	StatusArchived Status = "archived"
)

// Conversation 是会话聚合根（运行态）。
type Conversation struct {
	id           ConversationId
	title        string
	agentID      agent.AgentId
	agentVersion int
	turns        []*Turn
	summary      *ConversationSummary
	status       Status
	createdAt    time.Time
	updatedAt    time.Time
	events       []ddd.DomainEvent
}

// NewConversation 创建会话。
func NewConversation(id ConversationId, title string, agentID agent.AgentId, agentVersion int, now time.Time) (*Conversation, error) {
	if id.IsZero() {
		return nil, errors.New("conversation: id is required")
	}
	if agentID.IsZero() {
		return nil, errors.New("conversation: agent id is required")
	}
	if agentVersion <= 0 {
		return nil, errors.New("conversation: agent version must be positive")
	}
	if title == "" {
		title = "未命名会话"
	}
	c := &Conversation{
		id:           id,
		title:        title,
		agentID:      agentID,
		agentVersion: agentVersion,
		turns:        make([]*Turn, 0),
		status:       StatusActive,
		createdAt:    now,
		updatedAt:    now,
	}
	c.record(ConversationStarted{ConversationID: id, AgentID: agentID, AgentVersion: agentVersion, Occurred: now})
	return c, nil
}

// RehydrateConversation 从持久化状态重建会话。
func RehydrateConversation(
	id ConversationId, title string, agentID agent.AgentId, agentVersion int,
	turns []*Turn, summary *ConversationSummary, status Status, createdAt, updatedAt time.Time,
) *Conversation {
	if turns == nil {
		turns = make([]*Turn, 0)
	}
	return &Conversation{
		id: id, title: title, agentID: agentID, agentVersion: agentVersion, turns: turns,
		summary: summary, status: status, createdAt: createdAt, updatedAt: updatedAt,
	}
}

// ID 返回聚合标识。
func (c *Conversation) ID() ConversationId { return c.id }

// Title 返回标题。
func (c *Conversation) Title() string { return c.title }

// AgentID 返回绑定的智能体。
func (c *Conversation) AgentID() agent.AgentId { return c.agentID }

// AgentVersion 返回创建会话时的智能体版本。
func (c *Conversation) AgentVersion() int { return c.agentVersion }

// Turns 返回轮次列表。
func (c *Conversation) Turns() []*Turn { return c.turns }

// Summary 返回压缩摘要。
func (c *Conversation) Summary() *ConversationSummary { return c.summary }

// Status 返回状态。
func (c *Conversation) Status() Status { return c.status }

// CreatedAt 返回创建时间。
func (c *Conversation) CreatedAt() time.Time { return c.createdAt }

// UpdatedAt 返回更新时间。
func (c *Conversation) UpdatedAt() time.Time { return c.updatedAt }

// EnsureActive 校验会话处于活跃状态。
func (c *Conversation) EnsureActive() error {
	if c.status != StatusActive {
		return fmt.Errorf("conversation: %s is %s", c.id, c.status)
	}
	return nil
}

// StartTurn 开启一轮对话，返回可变实体供本轮追加消息。
func (c *Conversation) StartTurn(ids idgen.Generator, now time.Time) (*Turn, error) {
	if err := c.EnsureActive(); err != nil {
		return nil, err
	}
	if cur := c.CurrentTurn(); cur != nil && cur.Status() == TurnRunning {
		return nil, errors.New("conversation: previous turn is still running")
	}
	t := NewTurn(TurnId(ids.New()), len(c.turns), c.agentID, c.agentVersion, now)
	c.turns = append(c.turns, t)
	c.updatedAt = now
	return t, nil
}

// CurrentTurn 返回最后一个轮次，没有则返回 nil。
func (c *Conversation) CurrentTurn() *Turn {
	if len(c.turns) == 0 {
		return nil
	}
	return c.turns[len(c.turns)-1]
}

// LatestTurn 返回最后一个已完成且带有答复的轮次，供上下文压缩与展示。
func (c *Conversation) LatestTurn() *Turn { return c.CurrentTurn() }

// RecentMessages 返回最近 limit 条消息（跨轮次，按时间顺序）。
func (c *Conversation) RecentMessages(limit int) []Message {
	all := make([]Message, 0)
	for _, t := range c.turns {
		all = append(all, t.messages...)
	}
	if limit <= 0 || limit > len(all) {
		return all
	}
	return all[len(all)-limit:]
}

// CompleteTurn 标记当前轮次完成并记录事件。
func (c *Conversation) CompleteTurn(now time.Time) error {
	cur := c.CurrentTurn()
	if cur == nil {
		return errors.New("conversation: no turn to complete")
	}
	cur.MarkCompleted(now)
	c.updatedAt = now
	c.record(TurnFinished{ConversationID: c.id, TurnID: cur.ID(), Occurred: now})
	return nil
}

// FailTurn 标记当前轮次失败。
func (c *Conversation) FailTurn(reason string, now time.Time) {
	if cur := c.CurrentTurn(); cur != nil {
		cur.MarkFailed(reason, now)
	}
	c.updatedAt = now
}

// SetSummary 更新压缩摘要并记录事件。
func (c *Conversation) SetSummary(s *ConversationSummary, now time.Time) {
	c.summary = s
	c.updatedAt = now
	c.record(ContextCompressed{ConversationID: c.id, UpToTurnIndex: s.UpToTurnIndex(), Occurred: now})
}

// Archive 归档会话。
func (c *Conversation) Archive(now time.Time) {
	c.status = StatusArchived
	c.updatedAt = now
}

// PullEvents 返回并清空领域事件。
func (c *Conversation) PullEvents() []ddd.DomainEvent {
	events := c.events
	c.events = nil
	return events
}

func (c *Conversation) record(e ddd.DomainEvent) { c.events = append(c.events, e) }
