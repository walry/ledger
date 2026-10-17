package tool

import (
	"context"
	"errors"
)

// ErrNotFound 表示工具不存在。
var ErrNotFound = errors.New("tool: not found")

// Repository 是工具聚合的仓储接口。
type Repository interface {
	Save(ctx context.Context, t *Tool) error
	FindByID(ctx context.Context, id ToolId) (*Tool, error)
	FindByName(ctx context.Context, name string) (*Tool, error)
	// FindByIDs 批量按 ID 查询，返回顺序不保证，缺失的 ID 不报错。
	FindByIDs(ctx context.Context, ids []ToolId) ([]*Tool, error)
	List(ctx context.Context, offset, limit int) ([]*Tool, int64, error)
	Delete(ctx context.Context, id ToolId) error
}
