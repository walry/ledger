package report

import (
	"context"
	"errors"
)

// ErrNotFound 表示报告不存在。
var ErrNotFound = errors.New("report: not found")

// ListQuery 描述报告列表查询。
type ListQuery struct {
	Status Status
	Offset int
	Limit  int
}

// Repository 是报告聚合的仓储接口（实现在基础设施层）。
type Repository interface {
	Save(ctx context.Context, r *Report) error
	FindByID(ctx context.Context, id ReportId) (*Report, error)
	List(ctx context.Context, q ListQuery) ([]*Report, int64, error)
	Delete(ctx context.Context, id ReportId) error
}
