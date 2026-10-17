package analysis

import (
	"errors"
	"strings"
	"time"

	"github.com/walry/ledger/internal/domain/agent"
	"github.com/walry/ledger/internal/domain/conversation"
	"github.com/walry/ledger/internal/domain/ddd"
	"github.com/walry/ledger/internal/domain/report"
)

// TaskStatus 是分析任务状态。
type TaskStatus string

const (
	// TaskQueued 排队中。
	TaskQueued TaskStatus = "queued"
	// TaskRunning 执行中。
	TaskRunning TaskStatus = "running"
	// TaskSucceeded 成功。
	TaskSucceeded TaskStatus = "succeeded"
	// TaskFailed 失败。
	TaskFailed TaskStatus = "failed"
)

// AnalysisTask 是"数字员工"分析任务聚合根（运行态）。
type AnalysisTask struct {
	id             AnalysisTaskId
	title          string
	question       string
	agentID        agent.AgentId
	agentVersion   int
	reportIDs      []report.ReportId
	status         TaskStatus
	result         string
	failure        string
	conversationID conversation.ConversationId
	createdAt      time.Time
	startedAt      *time.Time
	finishedAt     *time.Time
	events         []ddd.DomainEvent
}

// NewAnalysisTask 创建分析任务。
func NewAnalysisTask(id AnalysisTaskId, title, question string, agentID agent.AgentId, agentVersion int, reportIDs []report.ReportId, now time.Time) (*AnalysisTask, error) {
	if id.IsZero() {
		return nil, errors.New("analysis: id is required")
	}
	if agentID.IsZero() {
		return nil, errors.New("analysis: agent id is required")
	}
	if strings.TrimSpace(question) == "" {
		return nil, errors.New("analysis: question is required")
	}
	if title == "" {
		title = "财务分析任务"
	}
	t := &AnalysisTask{
		id:           id,
		title:        title,
		question:     question,
		agentID:      agentID,
		agentVersion: agentVersion,
		reportIDs:    append([]report.ReportId{}, reportIDs...),
		status:       TaskQueued,
		createdAt:    now,
	}
	t.record(TaskSubmitted{TaskID: id, AgentID: agentID.String(), Occurred: now})
	return t, nil
}

// RehydrateAnalysisTask 从持久化状态重建任务。
func RehydrateAnalysisTask(
	id AnalysisTaskId, title, question string, agentID agent.AgentId, agentVersion int,
	reportIDs []report.ReportId, status TaskStatus, result, failure string,
	conversationID conversation.ConversationId, createdAt time.Time, startedAt, finishedAt *time.Time,
) *AnalysisTask {
	if reportIDs == nil {
		reportIDs = []report.ReportId{}
	}
	return &AnalysisTask{
		id: id, title: title, question: question, agentID: agentID, agentVersion: agentVersion,
		reportIDs: reportIDs, status: status, result: result, failure: failure,
		conversationID: conversationID, createdAt: createdAt, startedAt: startedAt, finishedAt: finishedAt,
	}
}

// ID 返回聚合标识。
func (t *AnalysisTask) ID() AnalysisTaskId { return t.id }

// Title 返回标题。
func (t *AnalysisTask) Title() string { return t.title }

// Question 返回分析问题。
func (t *AnalysisTask) Question() string { return t.question }

// AgentID 返回执行的智能体。
func (t *AnalysisTask) AgentID() agent.AgentId { return t.agentID }

// AgentVersion 返回智能体版本。
func (t *AnalysisTask) AgentVersion() int { return t.agentVersion }

// ReportIDs 返回关联报告 ID。
func (t *AnalysisTask) ReportIDs() []report.ReportId {
	out := make([]report.ReportId, len(t.reportIDs))
	copy(out, t.reportIDs)
	return out
}

// Status 返回状态。
func (t *AnalysisTask) Status() TaskStatus { return t.status }

// Result 返回分析结果。
func (t *AnalysisTask) Result() string { return t.result }

// Failure 返回失败原因。
func (t *AnalysisTask) Failure() string { return t.failure }

// ConversationID 返回关联会话。
func (t *AnalysisTask) ConversationID() conversation.ConversationId { return t.conversationID }

// CreatedAt 返回创建时间。
func (t *AnalysisTask) CreatedAt() time.Time { return t.createdAt }

// StartedAt 返回开始时间。
func (t *AnalysisTask) StartedAt() *time.Time { return t.startedAt }

// FinishedAt 返回结束时间。
func (t *AnalysisTask) FinishedAt() *time.Time { return t.finishedAt }

// AttachConversation 绑定执行会话。
func (t *AnalysisTask) AttachConversation(id conversation.ConversationId) { t.conversationID = id }

// MarkRunning 置为执行中。
func (t *AnalysisTask) MarkRunning(now time.Time) {
	t.status = TaskRunning
	t.startedAt = &now
}

// MarkSucceeded 置为成功。
func (t *AnalysisTask) MarkSucceeded(result string, now time.Time) {
	t.status = TaskSucceeded
	t.result = result
	t.failure = ""
	t.finishedAt = &now
	t.record(TaskCompleted{TaskID: t.id, Occurred: now})
}

// MarkFailed 置为失败。
func (t *AnalysisTask) MarkFailed(reason string, now time.Time) {
	t.status = TaskFailed
	t.failure = reason
	t.finishedAt = &now
}

// PullEvents 返回并清空领域事件。
func (t *AnalysisTask) PullEvents() []ddd.DomainEvent {
	events := t.events
	t.events = nil
	return events
}

func (t *AnalysisTask) record(e ddd.DomainEvent) { t.events = append(t.events, e) }
