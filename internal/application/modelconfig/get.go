package modelconfig

import (
	"context"
	"errors"
	"fmt"

	modelconfigdomain "github.com/walry/ledger/internal/domain/modelconfig"
)

// GetModelConfigUseCase 查询单个模型配置。
type GetModelConfigUseCase struct {
	repo modelconfigdomain.Repository
}

// NewGetModelConfigUseCase 创建模型配置详情用例。
func NewGetModelConfigUseCase(repo modelconfigdomain.Repository) *GetModelConfigUseCase {
	return &GetModelConfigUseCase{repo: repo}
}

// Execute 执行查询。
func (uc *GetModelConfigUseCase) Execute(ctx context.Context, cmd GetModelConfigCommand) (ModelConfigDTO, error) {
	mc, err := uc.repo.FindByID(ctx, modelconfigdomain.NewModelConfigId(cmd.ConfigID))
	if err != nil {
		if errors.Is(err, modelconfigdomain.ErrNotFound) {
			return ModelConfigDTO{}, fmt.Errorf("modelconfig: %s not found: %w", cmd.ConfigID, err)
		}
		return ModelConfigDTO{}, fmt.Errorf("modelconfig: load config: %w", err)
	}
	return toModelConfigDTO(mc), nil
}
