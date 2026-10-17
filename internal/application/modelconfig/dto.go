package modelconfig

import (
	"context"
	"time"

	"github.com/walry/ledger/internal/domain/ddd"
	modelconfigdomain "github.com/walry/ledger/internal/domain/modelconfig"
)

// CreateModelConfigCommand 是创建模型配置的输入。
type CreateModelConfigCommand struct {
	Name       string
	Provider   modelconfigdomain.ProviderType
	BaseURL    string
	APIKey     string
	Model      string
	Parameters modelconfigdomain.ModelParameters
}

// CreateModelConfigResult 是创建模型配置的输出。
type CreateModelConfigResult struct {
	ConfigID string
}

// UpdateModelConfigCommand 是更新模型配置的输入。
type UpdateModelConfigCommand struct {
	ConfigID   string
	BaseURL    string
	APIKey     string
	Model      string
	Parameters modelconfigdomain.ModelParameters
}

// UpdateModelConfigResult 是更新模型配置的输出。
type UpdateModelConfigResult struct {
	ConfigID string
}

// ListModelConfigsCommand 是分页查询模型配置的输入。
type ListModelConfigsCommand struct {
	Offset int
	Limit  int
}

// ListModelConfigsResult 是模型配置列表 DTO。
type ListModelConfigsResult struct {
	Configs []ModelConfigDTO
	Total   int64
}

// ModelConfigDTO 是模型配置 DTO。出于安全考虑不包含 APIKey。
type ModelConfigDTO struct {
	ConfigID   string
	Name       string
	Provider   string
	BaseURL    string
	Model      string
	Parameters modelconfigdomain.ModelParameters
	Enabled    bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// GetModelConfigCommand 是查询单个模型配置的输入。
type GetModelConfigCommand struct {
	ConfigID string
}

// publishEvents 发布聚合累积的领域事件，bus 为空时安全跳过。
func publishEvents(ctx context.Context, bus ddd.EventBus, events []ddd.DomainEvent) error {
	if bus == nil || len(events) == 0 {
		return nil
	}
	return bus.Publish(ctx, events...)
}
