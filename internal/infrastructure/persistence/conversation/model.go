// Package conversation provides the GORM persistence adapter for the
// conversation bounded context. It implements the domain port
// conversation.Repository.
package conversation

import (
	"encoding/json"
	"time"

	"github.com/walry/ledger/internal/domain/agent"
	domain "github.com/walry/ledger/internal/domain/conversation"
)

// ConversationRow is the persistence model for the Conversation aggregate root.
type ConversationRow struct {
	ID               string     `gorm:"column:id;primaryKey;type:text"`
	Title            string     `gorm:"column:title;type:text"`
	AgentID          string     `gorm:"column:agent_id;type:text;not null;index"`
	AgentVersion     int        `gorm:"column:agent_version;not null;default:0"`
	SummaryContent   string     `gorm:"column:summary_content;type:text"`
	SummaryUpTo      int        `gorm:"column:summary_up_to_index;not null;default:0"`
	SummaryCreatedAt *time.Time `gorm:"column:summary_created_at"`
	Status           string     `gorm:"column:status;type:text;not null;index"`
	CreatedAt        time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt        time.Time  `gorm:"column:updated_at;not null"`
}

// TableName pins the physical table name.
func (ConversationRow) TableName() string { return "conversations" }

// TurnRow is the persistence model for the Turn entity inside the Conversation
// aggregate. It is a child table with an ON DELETE CASCADE foreign key.
type TurnRow struct {
	ID             string     `gorm:"column:id;primaryKey;type:text"`
	ConversationID string     `gorm:"column:conversation_id;type:text;not null;index"`
	Index          int        `gorm:"column:idx;not null;default:0"`
	AgentID        string     `gorm:"column:agent_id;type:text"`
	AgentVersion   int        `gorm:"column:agent_version;not null;default:0"`
	Status         string     `gorm:"column:status;type:text;not null"`
	Failure        string     `gorm:"column:failure;type:text"`
	CreatedAt      time.Time  `gorm:"column:created_at;not null"`
	CompletedAt    *time.Time `gorm:"column:completed_at"`

	Conversation ConversationRow `gorm:"foreignKey:ConversationID;references:ID;constraint:OnDelete:CASCADE"`
}

// TableName pins the physical table name.
func (TurnRow) TableName() string { return "conversation_turns" }

// MessageRow is the persistence model for the Message value object inside a
// Turn. Messages have no natural id, so a surrogate bigserial is used.
type MessageRow struct {
	ID         uint64    `gorm:"column:id;primaryKey;autoIncrement"`
	TurnID     string    `gorm:"column:turn_id;type:text;not null;index"`
	Index      int       `gorm:"column:idx;not null;default:0"`
	Role       string    `gorm:"column:role;type:text;not null"`
	Content    string    `gorm:"column:content;type:text"`
	Name       string    `gorm:"column:name;type:text"`
	ToolCalls  []byte    `gorm:"column:tool_calls;type:jsonb"`
	ToolCallID string    `gorm:"column:tool_call_id;type:text"`
	CreatedAt  time.Time `gorm:"column:created_at;not null"`

	Turn TurnRow `gorm:"foreignKey:TurnID;references:ID;constraint:OnDelete:CASCADE"`
}

// TableName pins the physical table name.
func (MessageRow) TableName() string { return "conversation_messages" }

func marshalToolCalls(calls []domain.ToolCall) []byte {
	if calls == nil {
		calls = make([]domain.ToolCall, 0)
	}
	b, err := json.Marshal(calls)
	if err != nil {
		return []byte("[]")
	}
	return b
}

func unmarshalToolCalls(b []byte) []domain.ToolCall {
	out := make([]domain.ToolCall, 0)
	if len(b) == 0 {
		return out
	}
	if err := json.Unmarshal(b, &out); err != nil {
		return make([]domain.ToolCall, 0)
	}
	if out == nil {
		out = make([]domain.ToolCall, 0)
	}
	return out
}

func conversationToRow(c *domain.Conversation) *ConversationRow {
	row := &ConversationRow{
		ID:           c.ID().String(),
		Title:        c.Title(),
		AgentID:      c.AgentID().String(),
		AgentVersion: c.AgentVersion(),
		Status:       string(c.Status()),
		CreatedAt:    c.CreatedAt(),
		UpdatedAt:    c.UpdatedAt(),
	}
	if s := c.Summary(); s != nil {
		row.SummaryContent = s.Content()
		row.SummaryUpTo = s.UpToTurnIndex()
		createdAt := s.CreatedAt()
		row.SummaryCreatedAt = &createdAt
	}
	return row
}

func conversationFromRow(row *ConversationRow, turns []*domain.Turn) *domain.Conversation {
	var summary *domain.ConversationSummary
	if row.SummaryContent != "" {
		createdAt := row.CreatedAt
		if row.SummaryCreatedAt != nil {
			createdAt = *row.SummaryCreatedAt
		}
		if s, err := domain.NewConversationSummary(row.SummaryContent, row.SummaryUpTo, createdAt); err == nil {
			summary = s
		}
	}
	return domain.RehydrateConversation(
		domain.ConversationId(row.ID), row.Title, agent.AgentId(row.AgentID), row.AgentVersion,
		turns, summary, domain.Status(row.Status), row.CreatedAt, row.UpdatedAt,
	)
}

func turnToRow(conversationID string, t *domain.Turn) *TurnRow {
	return &TurnRow{
		ID:             t.ID().String(),
		ConversationID: conversationID,
		Index:          t.Index(),
		AgentID:        t.AgentID().String(),
		AgentVersion:   t.AgentVersion(),
		Status:         string(t.Status()),
		Failure:        t.Failure(),
		CreatedAt:      t.CreatedAt(),
		CompletedAt:    t.CompletedAt(),
	}
}

func turnFromRow(row *TurnRow, messages []domain.Message) *domain.Turn {
	return domain.RehydrateTurn(
		domain.TurnId(row.ID), row.Index, agent.AgentId(row.AgentID), row.AgentVersion,
		messages, domain.TurnStatus(row.Status), row.Failure, row.CreatedAt, row.CompletedAt,
	)
}

func messageToRow(turnID string, index int, m domain.Message) *MessageRow {
	return &MessageRow{
		TurnID:     turnID,
		Index:      index,
		Role:       string(m.Role()),
		Content:    m.Content(),
		Name:       m.Name(),
		ToolCalls:  marshalToolCalls(m.ToolCalls()),
		ToolCallID: m.ToolCallID(),
		CreatedAt:  m.CreatedAt(),
	}
}

func messageFromRow(row *MessageRow) domain.Message {
	return domain.RehydrateMessage(
		domain.Role(row.Role), row.Content, row.Name,
		unmarshalToolCalls(row.ToolCalls), row.ToolCallID, row.CreatedAt,
	)
}
