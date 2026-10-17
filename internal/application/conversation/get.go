package conversation

import (
	"context"
	"errors"
	"fmt"

	conversationdomain "github.com/walry/ledger/internal/domain/conversation"
)

// GetConversationUseCase 查询单个会话详情（含轮次与消息）。
type GetConversationUseCase struct {
	repo conversationdomain.Repository
}

// NewGetConversationUseCase 创建会话详情用例。
func NewGetConversationUseCase(repo conversationdomain.Repository) *GetConversationUseCase {
	return &GetConversationUseCase{repo: repo}
}

// Execute 执行查询。
func (uc *GetConversationUseCase) Execute(ctx context.Context, cmd GetConversationCommand) (ConversationDetailDTO, error) {
	c, err := uc.repo.FindByID(ctx, conversationdomain.NewConversationId(cmd.ConversationID))
	if err != nil {
		if errors.Is(err, conversationdomain.ErrNotFound) {
			return ConversationDetailDTO{}, fmt.Errorf("conversation: %s not found: %w", cmd.ConversationID, err)
		}
		return ConversationDetailDTO{}, fmt.Errorf("conversation: load conversation: %w", err)
	}

	turns := make([]TurnDTO, 0, len(c.Turns()))
	for _, t := range c.Turns() {
		messages := make([]MessageDTO, 0, len(t.Messages()))
		for _, m := range t.Messages() {
			calls := make([]ToolCallDTO, 0, len(m.ToolCalls()))
			for _, call := range m.ToolCalls() {
				calls = append(calls, ToolCallDTO{ID: call.ID, Name: call.Name, Arguments: call.Arguments})
			}
			messages = append(messages, MessageDTO{
				Role:      string(m.Role()),
				Content:   m.Content(),
				Name:      m.Name(),
				ToolCalls: calls,
				CreatedAt: m.CreatedAt(),
			})
		}
		turns = append(turns, TurnDTO{
			TurnID:       t.ID().String(),
			Index:        t.Index(),
			AgentVersion: t.AgentVersion(),
			Status:       string(t.Status()),
			Failure:      t.Failure(),
			CreatedAt:    t.CreatedAt(),
			Messages:     messages,
		})
	}

	var summary string
	if s := c.Summary(); s != nil {
		summary = s.Content()
	}

	return ConversationDetailDTO{
		ConversationID: c.ID().String(),
		Title:          c.Title(),
		AgentID:        c.AgentID().String(),
		AgentVersion:   c.AgentVersion(),
		Status:         string(c.Status()),
		Summary:        summary,
		CreatedAt:      c.CreatedAt(),
		UpdatedAt:      c.UpdatedAt(),
		Turns:          turns,
	}, nil
}
