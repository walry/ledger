package analysis

import (
	"time"

	"github.com/walry/ledger/internal/domain/ddd"
)

// TaskSubmitted 分析任务已提交。
type TaskSubmitted struct {
	TaskID   AnalysisTaskId
	AgentID  string
	Occurred time.Time
}

// EventName 实现 ddd.DomainEvent。
func (TaskSubmitted) EventName() string { return "AnalysisTaskSubmitted" }

// OccurredAt 实现 ddd.DomainEvent。
func (e TaskSubmitted) OccurredAt() time.Time { return e.Occurred }

// TaskCompleted 分析任务完成。
type TaskCompleted struct {
	TaskID   AnalysisTaskId
	Occurred time.Time
}

// EventName 实现 ddd.DomainEvent。
func (TaskCompleted) EventName() string { return "AnalysisTaskCompleted" }

// OccurredAt 实现 ddd.DomainEvent。
func (e TaskCompleted) OccurredAt() time.Time { return e.Occurred }

var _ ddd.DomainEvent = TaskSubmitted{}
var _ ddd.DomainEvent = TaskCompleted{}
