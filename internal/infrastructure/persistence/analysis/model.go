// Package analysis provides the GORM persistence adapter for the analysis
// bounded context ("数字员工"). It implements the domain port
// analysis.Repository.
package analysis

import (
	"encoding/json"
	"time"

	"github.com/walry/ledger/internal/domain/agent"
	domain "github.com/walry/ledger/internal/domain/analysis"
	domainconversation "github.com/walry/ledger/internal/domain/conversation"
	domainreport "github.com/walry/ledger/internal/domain/report"
)

// AnalysisTaskRow is the persistence model for the AnalysisTask aggregate.
type AnalysisTaskRow struct {
	ID             string     `gorm:"column:id;primaryKey;type:text"`
	Title          string     `gorm:"column:title;type:text"`
	Question       string     `gorm:"column:question;type:text;not null"`
	AgentID        string     `gorm:"column:agent_id;type:text;not null;index"`
	AgentVersion   int        `gorm:"column:agent_version;not null;default:0"`
	ReportIDs      []byte     `gorm:"column:report_ids;type:jsonb"`
	Status         string     `gorm:"column:status;type:text;not null;index:idx_analysis_tasks_status_created,priority:1"`
	Result         string     `gorm:"column:result;type:text"`
	Failure        string     `gorm:"column:failure;type:text"`
	ConversationID string     `gorm:"column:conversation_id;type:text"`
	CreatedAt      time.Time  `gorm:"column:created_at;not null;index:idx_analysis_tasks_status_created,priority:2"`
	StartedAt      *time.Time `gorm:"column:started_at"`
	FinishedAt     *time.Time `gorm:"column:finished_at"`
}

// TableName pins the physical table name.
func (AnalysisTaskRow) TableName() string { return "analysis_tasks" }

func reportIDsToJSON(ids []domainreport.ReportId) []byte {
	if ids == nil {
		ids = make([]domainreport.ReportId, 0)
	}
	b, err := json.Marshal(ids)
	if err != nil {
		return []byte("[]")
	}
	return b
}

func reportIDsFromJSON(b []byte) []domainreport.ReportId {
	out := make([]domainreport.ReportId, 0)
	if len(b) == 0 {
		return out
	}
	if err := json.Unmarshal(b, &out); err != nil {
		return make([]domainreport.ReportId, 0)
	}
	if out == nil {
		out = make([]domainreport.ReportId, 0)
	}
	return out
}

func analysisTaskToRow(t *domain.AnalysisTask) *AnalysisTaskRow {
	return &AnalysisTaskRow{
		ID:             t.ID().String(),
		Title:          t.Title(),
		Question:       t.Question(),
		AgentID:        t.AgentID().String(),
		AgentVersion:   t.AgentVersion(),
		ReportIDs:      reportIDsToJSON(t.ReportIDs()),
		Status:         string(t.Status()),
		Result:         t.Result(),
		Failure:        t.Failure(),
		ConversationID: t.ConversationID().String(),
		CreatedAt:      t.CreatedAt(),
		StartedAt:      t.StartedAt(),
		FinishedAt:     t.FinishedAt(),
	}
}

func analysisTaskFromRow(row *AnalysisTaskRow) *domain.AnalysisTask {
	return domain.RehydrateAnalysisTask(
		domain.AnalysisTaskId(row.ID), row.Title, row.Question, agent.AgentId(row.AgentID), row.AgentVersion,
		reportIDsFromJSON(row.ReportIDs), domain.TaskStatus(row.Status), row.Result, row.Failure,
		domainconversation.ConversationId(row.ConversationID), row.CreatedAt, row.StartedAt, row.FinishedAt,
	)
}
