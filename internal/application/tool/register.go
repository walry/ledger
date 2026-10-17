package tool

import (
	"context"
	"fmt"
	"time"

	"github.com/walry/ledger/internal/domain/ddd"
	tooldomain "github.com/walry/ledger/internal/domain/tool"
	"github.com/walry/ledger/pkg/clock"
	"github.com/walry/ledger/pkg/idgen"
)

// RegisterToolUseCase 注册工具：校验处理器存在后创建聚合。
type RegisterToolUseCase struct {
	repo    tooldomain.Repository
	handler tooldomain.HandlerFactory
	ids     idgen.Generator
	clk     clock.Clock
	bus     ddd.EventBus
}

// NewRegisterToolUseCase 创建注册工具用例。
func NewRegisterToolUseCase(repo tooldomain.Repository, handler tooldomain.HandlerFactory, ids idgen.Generator, clk clock.Clock, bus ddd.EventBus) *RegisterToolUseCase {
	return &RegisterToolUseCase{repo: repo, handler: handler, ids: ids, clk: clk, bus: bus}
}

// Execute 执行注册。
func (uc *RegisterToolUseCase) Execute(ctx context.Context, cmd RegisterToolCommand) (RegisterToolResult, error) {
	if uc.handler == nil || !uc.handler.Has(cmd.HandlerName) {
		return RegisterToolResult{}, fmt.Errorf("tool: handler %q not registered", cmd.HandlerName)
	}
	source, err := tooldomain.NewToolSource(cmd.SourceKind, cmd.SourceURI, cmd.SourceConfig)
	if err != nil {
		return RegisterToolResult{}, fmt.Errorf("tool: build source: %w", err)
	}
	handlerRef, err := tooldomain.NewHandlerRef(cmd.HandlerName)
	if err != nil {
		return RegisterToolResult{}, fmt.Errorf("tool: build handler ref: %w", err)
	}
	timeout := time.Duration(cmd.TimeoutSeconds) * time.Second

	t, err := tooldomain.NewTool(
		tooldomain.ToolId(uc.ids.New()),
		cmd.Name,
		cmd.Description,
		cmd.Schema,
		source,
		handlerRef,
		timeout,
		uc.clk.Now(),
	)
	if err != nil {
		return RegisterToolResult{}, fmt.Errorf("tool: create tool: %w", err)
	}
	if err := uc.repo.Save(ctx, t); err != nil {
		return RegisterToolResult{}, fmt.Errorf("tool: save tool: %w", err)
	}
	if err := publishEvents(ctx, uc.bus, t.PullEvents()); err != nil {
		return RegisterToolResult{}, fmt.Errorf("tool: publish events: %w", err)
	}
	return RegisterToolResult{ToolID: t.ID().String()}, nil
}
