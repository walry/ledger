package agent

import (
	"context"
	"errors"
	"fmt"

	agentdomain "github.com/walry/ledger/internal/domain/agent"
	"github.com/walry/ledger/internal/domain/ddd"
)

// RollbackUseCase 回滚到目标历史版本：以其为蓝本派生一个新的草稿版本。
type RollbackUseCase struct {
	repo    agentdomain.Repository
	factory *agentdomain.Factory
	bus     ddd.EventBus
}

// NewRollbackUseCase 创建回滚用例。
func NewRollbackUseCase(repo agentdomain.Repository, factory *agentdomain.Factory, bus ddd.EventBus) *RollbackUseCase {
	return &RollbackUseCase{repo: repo, factory: factory, bus: bus}
}

// Execute 执行回滚。
func (uc *RollbackUseCase) Execute(ctx context.Context, cmd RollbackCommand) (RollbackResult, error) {
	target, err := uc.repo.FindByID(ctx, agentdomain.NewAgentId(cmd.AgentID))
	if err != nil {
		if errors.Is(err, agentdomain.ErrNotFound) {
			return RollbackResult{}, fmt.Errorf("agent: %s not found: %w", cmd.AgentID, err)
		}
		return RollbackResult{}, fmt.Errorf("agent: load target version: %w", err)
	}
	next, err := uc.factory.NewVersion(target)
	if err != nil {
		return RollbackResult{}, fmt.Errorf("agent: derive rollback version: %w", err)
	}
	if err := uc.repo.Save(ctx, next); err != nil {
		return RollbackResult{}, fmt.Errorf("agent: save rollback version: %w", err)
	}
	if err := publishEvents(ctx, uc.bus, next.PullEvents()); err != nil {
		return RollbackResult{}, fmt.Errorf("agent: publish events: %w", err)
	}
	return RollbackResult{AgentID: next.ID().String(), Version: next.Version(), ParentVersion: next.ParentVersion()}, nil
}
