package runtime

import (
	"context"
	"fmt"
	"strings"

	"github.com/walry/ledger/internal/domain/analysis"
	"github.com/walry/ledger/internal/domain/conversation"
	"github.com/walry/ledger/internal/domain/ddd"
	"github.com/walry/ledger/internal/domain/report"
	"github.com/walry/ledger/pkg/clock"
	"github.com/walry/ledger/pkg/idgen"
)

// DigitalEmployeeConfig 是数字员工依赖。
type DigitalEmployeeConfig struct {
	Executor      *AgentExecutor
	Tasks         analysis.Repository
	Reports       report.Repository
	Conversations conversation.Repository
	Events        ddd.EventBus
	IDs           idgen.Generator
	Clock         clock.Clock
	// MaxReportChars 限制拼入提示词的报告正文长度。
	MaxReportChars int
}

// DigitalEmployee 是"数字员工"业务角色对应的领域服务：
// 读取分析任务，组装报告上下文，驱动 Agent 执行并回写结果。
type DigitalEmployee struct {
	cfg DigitalEmployeeConfig
}

// NewDigitalEmployee 创建数字员工。
func NewDigitalEmployee(cfg DigitalEmployeeConfig) *DigitalEmployee {
	if cfg.Clock == nil {
		cfg.Clock = clock.System()
	}
	if cfg.MaxReportChars <= 0 {
		cfg.MaxReportChars = 12000
	}
	return &DigitalEmployee{cfg: cfg}
}

// Execute 执行一个已入库的分析任务。
func (d *DigitalEmployee) Execute(ctx context.Context, taskID analysis.AnalysisTaskId) error {
	task, err := d.cfg.Tasks.FindByID(ctx, taskID)
	if err != nil {
		return err
	}
	now := d.cfg.Clock.Now()
	task.MarkRunning(now)

	conv, err := conversation.NewConversation(
		conversation.ConversationId(d.cfg.IDs.New()),
		task.Title(),
		task.AgentID(),
		task.AgentVersion(),
		now,
	)
	if err != nil {
		task.MarkFailed(err.Error(), d.cfg.Clock.Now())
		_ = d.persist(ctx, task)
		return err
	}
	task.AttachConversation(conv.ID())
	if err := d.cfg.Conversations.Save(ctx, conv); err != nil {
		task.MarkFailed(err.Error(), d.cfg.Clock.Now())
		_ = d.persist(ctx, task)
		return err
	}

	question := d.composeQuestion(ctx, task)
	result, err := d.cfg.Executor.Run(ctx, RunInput{ConversationID: conv.ID(), Question: question})
	if err != nil {
		task.MarkFailed(err.Error(), d.cfg.Clock.Now())
		_ = d.persist(ctx, task)
		return err
	}
	task.MarkSucceeded(result.Answer, d.cfg.Clock.Now())
	return d.persist(ctx, task)
}

// composeQuestion 将任务问题与报告正文拼接为执行问题。
func (d *DigitalEmployee) composeQuestion(ctx context.Context, task *analysis.AnalysisTask) string {
	var b strings.Builder
	b.WriteString(task.Question())
	b.WriteString("\n\n")
	for _, rid := range task.ReportIDs() {
		r, err := d.cfg.Reports.FindByID(ctx, rid)
		if err != nil {
			continue
		}
		text := r.RawText()
		if text == "" {
			for _, s := range r.Sections() {
				text += s.Text() + "\n"
			}
		}
		text = strings.TrimSpace(text)
		if text == "" {
			continue
		}
		if len(text) > d.cfg.MaxReportChars {
			text = text[:d.cfg.MaxReportChars] + "\n...(报告内容已截断)"
		}
		fmt.Fprintf(&b, "=== 报告《%s》(ID: %s) ===\n%s\n\n", r.Name(), r.ID(), text)
	}
	return b.String()
}

func (d *DigitalEmployee) persist(ctx context.Context, task *analysis.AnalysisTask) error {
	if err := d.cfg.Tasks.Save(ctx, task); err != nil {
		return err
	}
	if d.cfg.Events != nil {
		if events := task.PullEvents(); len(events) > 0 {
			return d.cfg.Events.Publish(ctx, events...)
		}
	}
	return nil
}
