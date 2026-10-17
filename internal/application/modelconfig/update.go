package modelconfig

import (
	"context"
	"errors"
	"fmt"

	"github.com/walry/ledger/internal/domain/ddd"
	modelconfigdomain "github.com/walry/ledger/internal/domain/modelconfig"
	"github.com/walry/ledger/pkg/clock"
)

// UpdateModelConfigUseCase 更新模型连接配置。
type UpdateModelConfigUseCase struct {
	repo modelconfigdomain.Repository
	clk  clock.Clock
	bus  ddd.EventBus
}

// NewUpdateModelConfigUseCase 创建更新模型配置用例。
func NewUpdateModelConfigUseCase(repo modelconfigdomain.Repository, clk clock.Clock, bus ddd.EventBus) *UpdateModelConfigUseCase {
	return &UpdateModelConfigUseCase{repo: repo, clk: clk, bus: bus}
}

// Execute 执行更新。
func (uc *UpdateModelConfigUseCase) Execute(ctx context.Context, cmd UpdateModelConfigCommand) (UpdateModelConfigResult, error) {
	mc, err := uc.repo.FindByID(ctx, modelconfigdomain.NewModelConfigId(cmd.ConfigID))
	if err != nil {
		if errors.Is(err, modelconfigdomain.ErrNotFound) {
			return UpdateModelConfigResult{}, fmt.Errorf("modelconfig: %s not found: %w", cmd.ConfigID, err)
		}
		return UpdateModelConfigResult{}, fmt.Errorf("modelconfig: load model config: %w", err)
	}
	if err := mc.Update(cmd.BaseURL, cmd.APIKey, cmd.Model, cmd.Parameters, uc.clk.Now()); err != nil {
		return UpdateModelConfigResult{}, fmt.Errorf("modelconfig: update: %w", err)
	}
	if err := uc.repo.Save(ctx, mc); err != nil {
		return UpdateModelConfigResult{}, fmt.Errorf("modelconfig: save model config: %w", err)
	}
	if err := publishEvents(ctx, uc.bus, mc.PullEvents()); err != nil {
		return UpdateModelConfigResult{}, fmt.Errorf("modelconfig: publish events: %w", err)
	}
	return UpdateModelConfigResult{ConfigID: mc.ID().String()}, nil
}
