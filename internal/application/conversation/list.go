package conversation

import (
	"context"
	"fmt"

	agentdomain "github.com/walry/ledger/internal/domain/agent"
	conversationdomain "github.com/walry/ledger/internal/domain/conversation"
)

// ListConversationsUseCase 分页查询会话。
type ListConversationsUseCase struct {
	repo conversationdomain.Repository
}

// NewListConversationsUseCase 创建会话列表用例。
func NewListConversationsUseCase(repo conversationdomain.Repository) *ListConversationsUseCase {
	return &ListConversationsUseCase{repo: repo}
}

// Execute 执行分页查询；指定 AgentID 时按智能体过滤。
func (uc *ListConversationsUseCase) Execute(ctx context.Context, cmd ListConversationsCommand) (ListConversationsResult, error) {
	var (
		conversations []*conversationdomain.Conversation
		total         int64
		err           error
	)
	if cmd.AgentID != "" {
		conversations, total, err = uc.repo.ListByAgent(ctx, agentdomain.NewAgentId(cmd.AgentID), cmd.Offset, cmd.Limit)
	} else {
		conversations, total, err = uc.repo.List(ctx, cmd.Offset, cmd.Limit)
	}
	if err != nil {
		return ListConversationsResult{}, fmt.Errorf("conversation: list conversations: %w", err)
	}
	items := make([]ConversationSummaryDTO, 0, len(conversations))
	for _, c := range conversations {
		items = append(items, ConversationSummaryDTO{
			ConversationID: c.ID().String(),
			Title:          c.Title(),
			AgentID:        c.AgentID().String(),
			AgentVersion:   c.AgentVersion(),
			Status:         string(c.Status()),
			CreatedAt:      c.CreatedAt(),
		})
	}
	return ListConversationsResult{Conversations: items, Total: total}, nil
}
