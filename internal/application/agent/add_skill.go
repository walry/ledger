package agent

import (
	"context"
	"errors"
	"fmt"

	agentdomain "github.com/walry/ledger/internal/domain/agent"
	"github.com/walry/ledger/internal/domain/ddd"
	"github.com/walry/ledger/pkg/clock"
)

// AddSkillUseCase 为草稿智能体绑定技能。
type AddSkillUseCase struct {
	repo agentdomain.Repository
	clk  clock.Clock
	bus  ddd.EventBus
}

// NewAddSkillUseCase 创建绑定技能用例。
func NewAddSkillUseCase(repo agentdomain.Repository, clk clock.Clock, bus ddd.EventBus) *AddSkillUseCase {
	return &AddSkillUseCase{repo: repo, clk: clk, bus: bus}
}

// Execute 执行绑定。
func (uc *AddSkillUseCase) Execute(ctx context.Context, cmd AddSkillCommand) (AddSkillResult, error) {
	a, err := uc.repo.FindByID(ctx, agentdomain.NewAgentId(cmd.AgentID))
	if err != nil {
		if errors.Is(err, agentdomain.ErrNotFound) {
			return AddSkillResult{}, fmt.Errorf("agent: %s not found: %w", cmd.AgentID, err)
		}
		return AddSkillResult{}, fmt.Errorf("agent: load agent: %w", err)
	}
	binding, err := agentdomain.NewSkillBinding(cmd.SkillID, cmd.Version)
	if err != nil {
		return AddSkillResult{}, fmt.Errorf("agent: build skill binding: %w", err)
	}
	if err := a.AddSkill(binding, uc.clk.Now()); err != nil {
		return AddSkillResult{}, fmt.Errorf("agent: add skill: %w", err)
	}
	if err := uc.repo.Save(ctx, a); err != nil {
		return AddSkillResult{}, fmt.Errorf("agent: save agent: %w", err)
	}
	if err := publishEvents(ctx, uc.bus, a.PullEvents()); err != nil {
		return AddSkillResult{}, fmt.Errorf("agent: publish events: %w", err)
	}
	return AddSkillResult{AgentID: a.ID().String()}, nil
}
