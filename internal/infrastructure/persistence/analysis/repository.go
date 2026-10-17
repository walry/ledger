package analysis

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	domain "github.com/walry/ledger/internal/domain/analysis"
)

// Repository is the GORM-backed implementation of analysis.Repository.
type Repository struct {
	db *gorm.DB
}

// NewRepository constructs the repository.
func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

var _ domain.Repository = (*Repository)(nil)

// Save upserts the analysis task aggregate root.
func (r *Repository) Save(ctx context.Context, t *domain.AnalysisTask) error {
	if t == nil {
		return errors.New("analysis: nil aggregate")
	}
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{UpdateAll: true}).
		Create(analysisTaskToRow(t)).Error
}

// FindByID loads a task, returning domain.ErrNotFound when absent.
func (r *Repository) FindByID(ctx context.Context, id domain.AnalysisTaskId) (*domain.AnalysisTask, error) {
	var row AnalysisTaskRow
	err := r.db.WithContext(ctx).Where("id = ?", id.String()).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return analysisTaskFromRow(&row), nil
}

// ClaimNextQueued atomically claims the oldest queued task and marks it
// running. It uses SELECT ... FOR UPDATE SKIP LOCKED so concurrent workers do
// not contend. Returns domain.ErrNotFound when the queue is empty.
func (r *Repository) ClaimNextQueued(ctx context.Context) (*domain.AnalysisTask, error) {
	var claimed *domain.AnalysisTask
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var row AnalysisTaskRow
		err := tx.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
			Where("status = ?", string(domain.TaskQueued)).
			Order("created_at asc, id asc").
			First(&row).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.ErrNotFound
		}
		if err != nil {
			return err
		}

		now := time.Now()
		row.Status = string(domain.TaskRunning)
		startedAt := now
		row.StartedAt = &startedAt

		if err := tx.Model(&AnalysisTaskRow{}).
			Where("id = ?", row.ID).
			Updates(map[string]any{
				"status":     row.Status,
				"started_at": startedAt,
			}).Error; err != nil {
			return err
		}

		claimed = analysisTaskFromRow(&row)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return claimed, nil
}

// List returns tasks plus the total count.
func (r *Repository) List(ctx context.Context, offset, limit int) ([]*domain.AnalysisTask, int64, error) {
	scope := r.db.WithContext(ctx).Model(&AnalysisTaskRow{})

	var total int64
	if err := scope.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	tx := scope.Order("created_at desc, id asc")
	if offset > 0 {
		tx = tx.Offset(offset)
	}
	if limit > 0 {
		tx = tx.Limit(limit)
	}

	var rows []AnalysisTaskRow
	if err := tx.Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	items := make([]*domain.AnalysisTask, 0, len(rows))
	for i := range rows {
		items = append(items, analysisTaskFromRow(&rows[i]))
	}
	return items, total, nil
}

// Delete removes a task by id.
func (r *Repository) Delete(ctx context.Context, id domain.AnalysisTaskId) error {
	return r.db.WithContext(ctx).Where("id = ?", id.String()).Delete(&AnalysisTaskRow{}).Error
}
