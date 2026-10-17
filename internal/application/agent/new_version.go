package agent

import (
	"context"
	"errors"
	"fmt"

	agentdomain "github.com/walry/ledger/internal/domain/agent"
	"github.com/walry/ledger/internal/domain/ddd"
)

// NewVersionUseCase 基于现有智能体派生一个新版本草稿。
type NewVersionUseCase struct {
	repo    agentdomain.Repository
	factory *agentdomain.Factory
	bus     ddd.EventBus
}

// NewNewVersionUseCase 创建派生新版本用例。
func NewNewVersionUseCase(repo agentdomain.Repository, factory *agentdomain.Factory, bus ddd.EventBus) *NewVersionUseCase {
	return &NewVersionUseCase{repo: repo, factory: factory, bus: bus}
}

// Execute 执行派生。
func (uc *NewVersionUseCase) Execute(ctx context.Context, cmd NewVersionCommand) (NewVersionResult, error) {
	base, err := uc.repo.FindByID(ctx, agentdomain.NewAgentId(cmd.AgentID))
	if err != nil {
		if errors.Is(err, agentdomain.ErrNotFound) {
			return NewVersionResult{}, fmt.Errorf("agent: %s not found: %w", cmd.AgentID, err)
		}
		return NewVersionResult{}, fmt.Errorf("agent: load agent: %w", err)
	}
	next, err := uc.factory.NewVersion(base)
	if err != nil {
		return NewVersionResult{}, fmt.Errorf("agent: derive new version: %w", err)
	}
	if err := uc.repo.Save(ctx, next); err != nil {
		return NewVersionResult{}, fmt.Errorf("agent: save new version: %w", err)
	}
	if err := publishEvents(ctx, uc.bus, next.PullEvents()); err != nil {
		return NewVersionResult{}, fmt.Errorf("agent: publish events: %w", err)
	}
	return NewVersionResult{AgentID: next.ID().String(), Version: next.Version(), ParentVersion: next.ParentVersion()}, nil
}
