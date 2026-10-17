package analysis

import (
	"context"
	"errors"
)

// ErrNotFound 表示分析任务不存在。
var ErrNotFound = errors.New("analysis: not found")

// Repository 是分析任务聚合的仓储接口。
type Repository interface {
	Save(ctx context.Context, t *AnalysisTask) error
	FindByID(ctx context.Context, id AnalysisTaskId) (*AnalysisTask, error)
	// ClaimNextQueued 原子地取出下一个排队任务并置为执行中，无任务时返回 ErrNotFound。
	ClaimNextQueued(ctx context.Context) (*AnalysisTask, error)
	List(ctx context.Context, offset, limit int) ([]*AnalysisTask, int64, error)
	Delete(ctx context.Context, id AnalysisTaskId) error
}
