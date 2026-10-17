package analysis

import (
	"context"
	"time"

	"github.com/walry/ledger/internal/domain/ddd"
)

// SubmitTaskCommand 是提交分析任务的输入。
type SubmitTaskCommand struct {
	Title     string
	Question  string
	AgentID   string
	ReportIDs []string
}

// SubmitTaskResult 是提交分析任务的输出。
type SubmitTaskResult struct {
	TaskID string
}

// RunTaskCommand 是执行分析任务的输入。
type RunTaskCommand struct {
	TaskID string
}

// GetResultCommand 是查询分析结果的输入。
type GetResultCommand struct {
	TaskID string
}

// GetResultResult 是分析结果 DTO。
type GetResultResult struct {
	TaskID         string
	Title          string
	Status         string
	Result         string
	Failure        string
	ConversationID string
	ReportIDs      []string
}

// ListTasksCommand 是分页查询分析任务的输入。
type ListTasksCommand struct {
	Offset int
	Limit  int
}

// ListTasksResult 是分析任务列表 DTO。
type ListTasksResult struct {
	Tasks []AnalysisTaskSummaryDTO
	Total int64
}

// AnalysisTaskSummaryDTO 是分析任务列表项 DTO。
type AnalysisTaskSummaryDTO struct {
	TaskID         string
	Title          string
	Status         string
	ConversationID string
	CreatedAt      time.Time
}

// publishEvents 发布聚合累积的领域事件，bus 为空时安全跳过。
func publishEvents(ctx context.Context, bus ddd.EventBus, events []ddd.DomainEvent) error {
	if bus == nil || len(events) == 0 {
		return nil
	}
	return bus.Publish(ctx, events...)
}
