package agent

import (
	"context"
	"errors"
	"fmt"

	agentdomain "github.com/walry/ledger/internal/domain/agent"
	"github.com/walry/ledger/internal/domain/ddd"
	"github.com/walry/ledger/pkg/clock"
)

// PublishUseCase 发布智能体（草稿 → 已发布）。
type PublishUseCase struct {
	repo agentdomain.Repository
	clk  clock.Clock
	bus  ddd.EventBus
}

// NewPublishUseCase 创建发布智能体用例。
func NewPublishUseCase(repo agentdomain.Repository, clk clock.Clock, bus ddd.EventBus) *PublishUseCase {
	return &PublishUseCase{repo: repo, clk: clk, bus: bus}
}

// Execute 执行发布。
func (uc *PublishUseCase) Execute(ctx context.Context, cmd PublishCommand) (PublishResult, error) {
	a, err := uc.repo.FindByID(ctx, agentdomain.NewAgentId(cmd.AgentID))
	if err != nil {
		if errors.Is(err, agentdomain.ErrNotFound) {
			return PublishResult{}, fmt.Errorf("agent: %s not found: %w", cmd.AgentID, err)
		}
		return PublishResult{}, fmt.Errorf("agent: load agent: %w", err)
	}
	if err := a.Publish(uc.clk.Now()); err != nil {
		return PublishResult{}, fmt.Errorf("agent: publish: %w", err)
	}
	if err := uc.repo.Save(ctx, a); err != nil {
		return PublishResult{}, fmt.Errorf("agent: save agent: %w", err)
	}
	if err := publishEvents(ctx, uc.bus, a.PullEvents()); err != nil {
		return PublishResult{}, fmt.Errorf("agent: publish events: %w", err)
	}
	return PublishResult{AgentID: a.ID().String(), Version: a.Version()}, nil
}
