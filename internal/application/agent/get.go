package agent

import (
	"context"
	"errors"
	"fmt"

	agentdomain "github.com/walry/ledger/internal/domain/agent"
)

// GetAgentUseCase 查询单个智能体详情。
type GetAgentUseCase struct {
	repo agentdomain.Repository
}

// NewGetAgentUseCase 创建智能体详情用例。
func NewGetAgentUseCase(repo agentdomain.Repository) *GetAgentUseCase {
	return &GetAgentUseCase{repo: repo}
}

// Execute 执行查询。
func (uc *GetAgentUseCase) Execute(ctx context.Context, cmd GetAgentCommand) (AgentDetailDTO, error) {
	a, err := uc.repo.FindByID(ctx, agentdomain.NewAgentId(cmd.AgentID))
	if err != nil {
		if errors.Is(err, agentdomain.ErrNotFound) {
			return AgentDetailDTO{}, fmt.Errorf("agent: %s not found: %w", cmd.AgentID, err)
		}
		return AgentDetailDTO{}, fmt.Errorf("agent: load agent: %w", err)
	}
	ref := a.ModelRef()
	ctxCfg := a.Context()

	kbIDs := make([]string, 0, len(a.KnowledgeBaseIDs()))
	for _, id := range a.KnowledgeBaseIDs() {
		kbIDs = append(kbIDs, id.String())
	}
	skills := make([]SkillBindingDTO, 0, len(a.Skills()))
	for _, b := range a.Skills() {
		skills = append(skills, SkillBindingDTO{SkillID: b.SkillID, Version: b.Version, Enabled: b.Enabled})
	}
	tools := make([]ToolBindingDTO, 0, len(a.Tools()))
	for _, b := range a.Tools() {
		tools = append(tools, ToolBindingDTO{
			ToolID: b.ToolID.String(), Version: b.Version, Enabled: b.Enabled, Config: b.Config,
		})
	}

	return AgentDetailDTO{
		AgentID:          a.ID().String(),
		Name:             a.Name(),
		Description:      a.Description(),
		Status:           a.Status().String(),
		SystemPrompt:     a.SystemPrompt().Template(),
		ModelConfigID:    ref.ConfigID.String(),
		Model:            ref.Model,
		Provider:         string(ref.Provider),
		KnowledgeBaseIDs: kbIDs,
		MaxSteps:         ctxCfg.MaxSteps(),
		RetrievalTopK:    ctxCfg.RetrievalTopK(),
		Version:          a.Version(),
		ParentVersion:    a.ParentVersion(),
		CreatedAt:        a.CreatedAt(),
		UpdatedAt:        a.UpdatedAt(),
		PublishedAt:      a.PublishedAt(),
		Skills:           skills,
		Tools:            tools,
	}, nil
}
