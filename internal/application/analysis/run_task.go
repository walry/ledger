package analysis

import (
	"context"
	"fmt"

	analysisdomain "github.com/walry/ledger/internal/domain/analysis"
	runtimedomain "github.com/walry/ledger/internal/domain/runtime"
)

// RunTaskUseCase 执行分析任务：编排 runtime.DigitalEmployee。
type RunTaskUseCase struct {
	employee *runtimedomain.DigitalEmployee
}

// NewRunTaskUseCase 创建执行分析任务用例。
func NewRunTaskUseCase(employee *runtimedomain.DigitalEmployee) *RunTaskUseCase {
	return &RunTaskUseCase{employee: employee}
}

// Execute 执行任务。
func (uc *RunTaskUseCase) Execute(ctx context.Context, cmd RunTaskCommand) error {
	if err := uc.employee.Execute(ctx, analysisdomain.NewAnalysisTaskId(cmd.TaskID)); err != nil {
		return fmt.Errorf("analysis: run task %s: %w", cmd.TaskID, err)
	}
	return nil
}
