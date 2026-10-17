// Package job 承载后台任务（如数字员工分析任务的异步执行）。
package job

import (
	"context"
	"errors"
	"log/slog"
	"time"

	analysisapp "github.com/walry/ledger/internal/application/analysis"
	analysisdomain "github.com/walry/ledger/internal/domain/analysis"
)

// AnalysisWorker 轮询分析任务队列并执行。
type AnalysisWorker struct {
	run      *analysisapp.RunTaskUseCase
	tasks    analysisdomain.Repository
	logger   *slog.Logger
	interval time.Duration
}

// NewAnalysisWorker 创建工作者。
func NewAnalysisWorker(run *analysisapp.RunTaskUseCase, tasks analysisdomain.Repository, logger *slog.Logger) *AnalysisWorker {
	if logger == nil {
		logger = slog.Default()
	}
	return &AnalysisWorker{run: run, tasks: tasks, logger: logger, interval: 2 * time.Second}
}

// Start 启动 workers 个并发工作者，直到 ctx 取消。
func (w *AnalysisWorker) Start(ctx context.Context, workers int) {
	if workers <= 0 {
		workers = 1
	}
	for i := 0; i < workers; i++ {
		go w.loop(ctx, i)
	}
}

func (w *AnalysisWorker) loop(ctx context.Context, id int) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.drain(ctx, id)
		}
	}
}

func (w *AnalysisWorker) drain(ctx context.Context, id int) {
	for {
		task, err := w.tasks.ClaimNextQueued(ctx)
		if err != nil {
			if !errors.Is(err, analysisdomain.ErrNotFound) {
				w.logger.Error("claim analysis task", "worker", id, "error", err)
			}
			return
		}
		taskID := task.ID().String()
		w.logger.Info("analysis task started", "worker", id, "taskId", taskID)
		if err := w.run.Execute(ctx, analysisapp.RunTaskCommand{TaskID: taskID}); err != nil {
			w.logger.Error("analysis task failed", "worker", id, "taskId", taskID, "error", err)
			continue
		}
		w.logger.Info("analysis task finished", "worker", id, "taskId", taskID)
	}
}
