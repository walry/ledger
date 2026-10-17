package tool

import (
	"context"
	"errors"
	"fmt"

	"github.com/walry/ledger/internal/domain/ddd"
	tooldomain "github.com/walry/ledger/internal/domain/tool"
	"github.com/walry/ledger/pkg/clock"
)

// EnableToolUseCase 启用或停用工具。
type EnableToolUseCase struct {
	repo tooldomain.Repository
	clk  clock.Clock
	bus  ddd.EventBus
}

// NewEnableToolUseCase 创建启用/停用工具用例。
func NewEnableToolUseCase(repo tooldomain.Repository, clk clock.Clock, bus ddd.EventBus) *EnableToolUseCase {
	return &EnableToolUseCase{repo: repo, clk: clk, bus: bus}
}

// Execute 执行启用/停用。
func (uc *EnableToolUseCase) Execute(ctx context.Context, cmd EnableToolCommand) (EnableToolResult, error) {
	t, err := uc.repo.FindByID(ctx, tooldomain.NewToolId(cmd.ToolID))
	if err != nil {
		if errors.Is(err, tooldomain.ErrNotFound) {
			return EnableToolResult{}, fmt.Errorf("tool: %s not found: %w", cmd.ToolID, err)
		}
		return EnableToolResult{}, fmt.Errorf("tool: load tool: %w", err)
	}
	now := uc.clk.Now()
	if cmd.Enabled {
		t.Enable(now)
	} else {
		t.Disable(now)
	}
	if err := uc.repo.Save(ctx, t); err != nil {
		return EnableToolResult{}, fmt.Errorf("tool: save tool: %w", err)
	}
	if err := publishEvents(ctx, uc.bus, t.PullEvents()); err != nil {
		return EnableToolResult{}, fmt.Errorf("tool: publish events: %w", err)
	}
	return EnableToolResult{ToolID: t.ID().String(), Enabled: t.Enabled()}, nil
}
