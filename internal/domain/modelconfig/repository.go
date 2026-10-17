package modelconfig

import (
	"context"
	"errors"
)

// ErrNotFound 表示模型配置不存在。
var ErrNotFound = errors.New("modelconfig: not found")

// Repository 是模型配置聚合的仓储接口。
type Repository interface {
	Save(ctx context.Context, m *ModelConfig) error
	FindByID(ctx context.Context, id ModelConfigId) (*ModelConfig, error)
	List(ctx context.Context, offset, limit int) ([]*ModelConfig, int64, error)
	Delete(ctx context.Context, id ModelConfigId) error
}
