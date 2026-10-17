package modelconfig

import (
	"context"
	"fmt"

	"github.com/walry/ledger/internal/domain/ddd"
	modelconfigdomain "github.com/walry/ledger/internal/domain/modelconfig"
	"github.com/walry/ledger/pkg/clock"
	"github.com/walry/ledger/pkg/idgen"
)

// CreateModelConfigUseCase 创建模型连接配置。
type CreateModelConfigUseCase struct {
	repo modelconfigdomain.Repository
	ids  idgen.Generator
	clk  clock.Clock
	bus  ddd.EventBus
}

// NewCreateModelConfigUseCase 创建模型配置用例。
func NewCreateModelConfigUseCase(repo modelconfigdomain.Repository, ids idgen.Generator, clk clock.Clock, bus ddd.EventBus) *CreateModelConfigUseCase {
	return &CreateModelConfigUseCase{repo: repo, ids: ids, clk: clk, bus: bus}
}

// Execute 执行创建。
func (uc *CreateModelConfigUseCase) Execute(ctx context.Context, cmd CreateModelConfigCommand) (CreateModelConfigResult, error) {
	mc, err := modelconfigdomain.NewModelConfig(
		modelconfigdomain.ModelConfigId(uc.ids.New()),
		cmd.Name,
		cmd.Provider,
		cmd.BaseURL,
		cmd.APIKey,
		cmd.Model,
		cmd.Parameters,
		uc.clk.Now(),
	)
	if err != nil {
		return CreateModelConfigResult{}, fmt.Errorf("modelconfig: create model config: %w", err)
	}
	if err := uc.repo.Save(ctx, mc); err != nil {
		return CreateModelConfigResult{}, fmt.Errorf("modelconfig: save model config: %w", err)
	}
	if err := publishEvents(ctx, uc.bus, mc.PullEvents()); err != nil {
		return CreateModelConfigResult{}, fmt.Errorf("modelconfig: publish events: %w", err)
	}
	return CreateModelConfigResult{ConfigID: mc.ID().String()}, nil
}
