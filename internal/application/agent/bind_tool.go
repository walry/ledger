package agent

import (
	"context"
	"errors"
	"fmt"

	agentdomain "github.com/walry/ledger/internal/domain/agent"
	"github.com/walry/ledger/internal/domain/ddd"
	tooldomain "github.com/walry/ledger/internal/domain/tool"
	"github.com/walry/ledger/pkg/clock"
)

// BindToolUseCase 为草稿智能体绑定工具（校验工具存在）。
type BindToolUseCase struct {
	repo  agentdomain.Repository
	tools tooldomain.Repository
	clk   clock.Clock
	bus   ddd.EventBus
}

// NewBindToolUseCase 创建绑定工具用例。
func NewBindToolUseCase(repo agentdomain.Repository, tools tooldomain.Repository, clk clock.Clock, bus ddd.EventBus) *BindToolUseCase {
	return &BindToolUseCase{repo: repo, tools: tools, clk: clk, bus: bus}
}

// Execute 执行绑定。
func (uc *BindToolUseCase) Execute(ctx context.Context, cmd BindToolCommand) (BindToolResult, error) {
	a, err := uc.repo.FindByID(ctx, agentdomain.NewAgentId(cmd.AgentID))
	if err != nil {
		if errors.Is(err, agentdomain.ErrNotFound) {
			return BindToolResult{}, fmt.Errorf("agent: %s not found: %w", cmd.AgentID, err)
		}
		return BindToolResult{}, fmt.Errorf("agent: load agent: %w", err)
	}

	t, err := uc.tools.FindByID(ctx, tooldomain.NewToolId(cmd.ToolID))
	if err != nil {
		if errors.Is(err, tooldomain.ErrNotFound) {
			return BindToolResult{}, fmt.Errorf("agent: tool %s not found: %w", cmd.ToolID, err)
		}
		return BindToolResult{}, fmt.Errorf("agent: load tool: %w", err)
	}

	binding, err := agentdomain.NewToolBinding(t.ID(), cmd.Version, cmd.Config)
	if err != nil {
		return BindToolResult{}, fmt.Errorf("agent: build tool binding: %w", err)
	}
	if err := a.BindTool(binding, uc.clk.Now()); err != nil {
		return BindToolResult{}, fmt.Errorf("agent: bind tool: %w", err)
	}
	if err := uc.repo.Save(ctx, a); err != nil {
		return BindToolResult{}, fmt.Errorf("agent: save agent: %w", err)
	}
	if err := publishEvents(ctx, uc.bus, a.PullEvents()); err != nil {
		return BindToolResult{}, fmt.Errorf("agent: publish events: %w", err)
	}
	return BindToolResult{AgentID: a.ID().String()}, nil
}
