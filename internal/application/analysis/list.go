package analysis

import (
	"context"
	"fmt"

	analysisdomain "github.com/walry/ledger/internal/domain/analysis"
)

// ListTasksUseCase 分页查询分析任务。
type ListTasksUseCase struct {
	tasks analysisdomain.Repository
}

// NewListTasksUseCase 创建分析任务列表用例。
func NewListTasksUseCase(tasks analysisdomain.Repository) *ListTasksUseCase {
	return &ListTasksUseCase{tasks: tasks}
}

// Execute 执行分页查询。
func (uc *ListTasksUseCase) Execute(ctx context.Context, cmd ListTasksCommand) (ListTasksResult, error) {
	tasks, total, err := uc.tasks.List(ctx, cmd.Offset, cmd.Limit)
	if err != nil {
		return ListTasksResult{}, fmt.Errorf("analysis: list tasks: %w", err)
	}
	items := make([]AnalysisTaskSummaryDTO, 0, len(tasks))
	for _, t := range tasks {
		items = append(items, AnalysisTaskSummaryDTO{
			TaskID:         t.ID().String(),
			Title:          t.Title(),
			Status:         string(t.Status()),
			ConversationID: t.ConversationID().String(),
			CreatedAt:      t.CreatedAt(),
		})
	}
	return ListTasksResult{Tasks: items, Total: total}, nil
}
