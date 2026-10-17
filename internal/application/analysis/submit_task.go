package analysis

import (
	"context"
	"errors"
	"fmt"

	agentdomain "github.com/walry/ledger/internal/domain/agent"
	analysisdomain "github.com/walry/ledger/internal/domain/analysis"
	"github.com/walry/ledger/internal/domain/ddd"
	reportdomain "github.com/walry/ledger/internal/domain/report"
	"github.com/walry/ledger/pkg/clock"
	"github.com/walry/ledger/pkg/idgen"
)

// SubmitTaskUseCase 提交"数字员工"分析任务（入队待执行）。
type SubmitTaskUseCase struct {
	tasks  analysisdomain.Repository
	agents agentdomain.Repository
	ids    idgen.Generator
	clk    clock.Clock
	bus    ddd.EventBus
}

// NewSubmitTaskUseCase 创建提交分析任务用例。
func NewSubmitTaskUseCase(tasks analysisdomain.Repository, agents agentdomain.Repository, ids idgen.Generator, clk clock.Clock, bus ddd.EventBus) *SubmitTaskUseCase {
	return &SubmitTaskUseCase{tasks: tasks, agents: agents, ids: ids, clk: clk, bus: bus}
}

// Execute 执行提交。
func (uc *SubmitTaskUseCase) Execute(ctx context.Context, cmd SubmitTaskCommand) (SubmitTaskResult, error) {
	a, err := uc.agents.FindByID(ctx, agentdomain.NewAgentId(cmd.AgentID))
	if err != nil {
		if errors.Is(err, agentdomain.ErrNotFound) {
			return SubmitTaskResult{}, fmt.Errorf("analysis: agent %s not found: %w", cmd.AgentID, err)
		}
		return SubmitTaskResult{}, fmt.Errorf("analysis: load agent: %w", err)
	}

	reportIDs := make([]reportdomain.ReportId, 0, len(cmd.ReportIDs))
	for _, id := range cmd.ReportIDs {
		reportIDs = append(reportIDs, reportdomain.NewReportId(id))
	}

	task, err := analysisdomain.NewAnalysisTask(
		analysisdomain.AnalysisTaskId(uc.ids.New()),
		cmd.Title,
		cmd.Question,
		a.ID(),
		a.Version(),
		reportIDs,
		uc.clk.Now(),
	)
	if err != nil {
		return SubmitTaskResult{}, fmt.Errorf("analysis: create task: %w", err)
	}
	if err := uc.tasks.Save(ctx, task); err != nil {
		return SubmitTaskResult{}, fmt.Errorf("analysis: save task: %w", err)
	}
	if err := publishEvents(ctx, uc.bus, task.PullEvents()); err != nil {
		return SubmitTaskResult{}, fmt.Errorf("analysis: publish events: %w", err)
	}
	return SubmitTaskResult{TaskID: task.ID().String()}, nil
}
