package analysis

import (
	"context"
	"errors"
	"fmt"

	analysisdomain "github.com/walry/ledger/internal/domain/analysis"
)

// GetResultUseCase 查询分析任务的状态与结果。
type GetResultUseCase struct {
	tasks analysisdomain.Repository
}

// NewGetResultUseCase 创建查询分析结果用例。
func NewGetResultUseCase(tasks analysisdomain.Repository) *GetResultUseCase {
	return &GetResultUseCase{tasks: tasks}
}

// Execute 执行查询。
func (uc *GetResultUseCase) Execute(ctx context.Context, cmd GetResultCommand) (GetResultResult, error) {
	task, err := uc.tasks.FindByID(ctx, analysisdomain.NewAnalysisTaskId(cmd.TaskID))
	if err != nil {
		if errors.Is(err, analysisdomain.ErrNotFound) {
			return GetResultResult{}, fmt.Errorf("analysis: task %s not found: %w", cmd.TaskID, err)
		}
		return GetResultResult{}, fmt.Errorf("analysis: load task: %w", err)
	}
	reportIDs := make([]string, 0, len(task.ReportIDs()))
	for _, id := range task.ReportIDs() {
		reportIDs = append(reportIDs, id.String())
	}
	return GetResultResult{
		TaskID:         task.ID().String(),
		Title:          task.Title(),
		Status:         string(task.Status()),
		Result:         task.Result(),
		Failure:        task.Failure(),
		ConversationID: task.ConversationID().String(),
		ReportIDs:      reportIDs,
	}, nil
}
