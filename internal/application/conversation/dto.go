package conversation

import (
	"context"
	"time"

	"github.com/walry/ledger/internal/domain/ddd"
)

// CreateConversationCommand 是创建会话的输入。
type CreateConversationCommand struct {
	AgentID string
	Title   string
}

// CreateConversationResult 是创建会话的输出。
type CreateConversationResult struct {
	ConversationID string
	AgentID        string
	AgentVersion   int
}

// StartTurnCommand 是开启一轮对话的输入。
type StartTurnCommand struct {
	ConversationID string
	Message        string
}

// StartTurnResult 是开启一轮对话的输出。
type StartTurnResult struct {
	ConversationID string
	TurnID         string
}

// CompressContextCommand 是压缩上下文的输入。
type CompressContextCommand struct {
	ConversationID string
}

// CompressContextResult 是压缩上下文的输出。
type CompressContextResult struct {
	ConversationID string
	Summary        string
	Compressed     bool
}

// ListConversationsCommand 是分页查询会话的输入。AgentID 非空时按智能体过滤。
type ListConversationsCommand struct {
	AgentID string
	Offset  int
	Limit   int
}

// ListConversationsResult 是会话列表 DTO。
type ListConversationsResult struct {
	Conversations []ConversationSummaryDTO
	Total         int64
}

// ConversationSummaryDTO 是会话列表项 DTO（不含轮次）。
type ConversationSummaryDTO struct {
	ConversationID string
	Title          string
	AgentID        string
	AgentVersion   int
	Status         string
	CreatedAt      time.Time
}

// GetConversationCommand 是查询单个会话的输入。
type GetConversationCommand struct {
	ConversationID string
}

// ConversationDetailDTO 是会话详情 DTO，含轮次与消息。
type ConversationDetailDTO struct {
	ConversationID string
	Title          string
	AgentID        string
	AgentVersion   int
	Status         string
	Summary        string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Turns          []TurnDTO
}

// TurnDTO 是轮次 DTO。
type TurnDTO struct {
	TurnID       string
	Index        int
	AgentVersion int
	Status       string
	Failure      string
	CreatedAt    time.Time
	Messages     []MessageDTO
}

// MessageDTO 是消息 DTO。
type MessageDTO struct {
	Role      string
	Content   string
	Name      string
	ToolCalls []ToolCallDTO
	CreatedAt time.Time
}

// ToolCallDTO 是工具调用 DTO。
type ToolCallDTO struct {
	ID        string
	Name      string
	Arguments string
}

// publishEvents 发布聚合累积的领域事件，bus 为空时安全跳过。
func publishEvents(ctx context.Context, bus ddd.EventBus, events []ddd.DomainEvent) error {
	if bus == nil || len(events) == 0 {
		return nil
	}
	return bus.Publish(ctx, events...)
}
