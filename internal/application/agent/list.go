package agent

import (
	"context"
	"fmt"

	agentdomain "github.com/walry/ledger/internal/domain/agent"
)

// ListAgentsUseCase 分页查询智能体。
type ListAgentsUseCase struct {
	repo agentdomain.Repository
}

// NewListAgentsUseCase 创建智能体列表用例。
func NewListAgentsUseCase(repo agentdomain.Repository) *ListAgentsUseCase {
	return &ListAgentsUseCase{repo: repo}
}

// Execute 执行分页查询。
func (uc *ListAgentsUseCase) Execute(ctx context.Context, cmd ListAgentsCommand) (ListAgentsResult, error) {
	agents, total, err := uc.repo.List(ctx, cmd.Offset, cmd.Limit)
	if err != nil {
		return ListAgentsResult{}, fmt.Errorf("agent: list agents: %w", err)
	}
	items := make([]AgentSummaryDTO, 0, len(agents))
	for _, a := range agents {
		items = append(items, toAgentSummary(a))
	}
	return ListAgentsResult{Agents: items, Total: total}, nil
}

// toAgentSummary 将聚合映射为列表项 DTO。
func toAgentSummary(a *agentdomain.Agent) AgentSummaryDTO {
	ref := a.ModelRef()
	return AgentSummaryDTO{
		AgentID:       a.ID().String(),
		Name:          a.Name(),
		Description:   a.Description(),
		Status:        a.Status().String(),
		Version:       a.Version(),
		ModelConfigID: ref.ConfigID.String(),
		Model:         ref.Model,
		Provider:      string(ref.Provider),
	}
}
