package modelconfig

import (
	"context"
	"fmt"

	modelconfigdomain "github.com/walry/ledger/internal/domain/modelconfig"
)

// ListModelConfigsUseCase 分页查询模型配置。
type ListModelConfigsUseCase struct {
	repo modelconfigdomain.Repository
}

// NewListModelConfigsUseCase 创建模型配置列表用例。
func NewListModelConfigsUseCase(repo modelconfigdomain.Repository) *ListModelConfigsUseCase {
	return &ListModelConfigsUseCase{repo: repo}
}

// Execute 执行分页查询。
func (uc *ListModelConfigsUseCase) Execute(ctx context.Context, cmd ListModelConfigsCommand) (ListModelConfigsResult, error) {
	configs, total, err := uc.repo.List(ctx, cmd.Offset, cmd.Limit)
	if err != nil {
		return ListModelConfigsResult{}, fmt.Errorf("modelconfig: list model configs: %w", err)
	}
	items := make([]ModelConfigDTO, 0, len(configs))
	for _, mc := range configs {
		items = append(items, toModelConfigDTO(mc))
	}
	return ListModelConfigsResult{Configs: items, Total: total}, nil
}

// toModelConfigDTO 映射为 DTO，且绝不携带 APIKey。
func toModelConfigDTO(mc *modelconfigdomain.ModelConfig) ModelConfigDTO {
	return ModelConfigDTO{
		ConfigID:   mc.ID().String(),
		Name:       mc.Name(),
		Provider:   string(mc.Provider()),
		BaseURL:    mc.BaseURL(),
		Model:      mc.Model(),
		Parameters: mc.Parameters(),
		Enabled:    mc.Enabled(),
		CreatedAt:  mc.CreatedAt(),
		UpdatedAt:  mc.UpdatedAt(),
	}
}
