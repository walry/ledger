package report

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	domain "github.com/walry/ledger/internal/domain/report"
)

// Repository is the GORM-backed implementation of report.Repository.
type Repository struct {
	db *gorm.DB
}

// NewRepository constructs the repository.
func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

var _ domain.Repository = (*Repository)(nil)

// Save upserts the aggregate root and replaces its sections inside one tx.
func (r *Repository) Save(ctx context.Context, agg *domain.Report) error {
	if agg == nil {
		return errors.New("report: nil aggregate")
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		root := reportToRow(agg)
		if err := tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(root).Error; err != nil {
			return err
		}
		if err := tx.Where("report_id = ?", root.ID).Delete(&SectionRow{}).Error; err != nil {
			return err
		}
		sections := agg.Sections()
		if len(sections) == 0 {
			return nil
		}
		rows := make([]SectionRow, 0, len(sections))
		for _, s := range sections {
			rows = append(rows, *sectionToRow(root.ID, s))
		}
		return tx.Create(&rows).Error
	})
}

// FindByID loads an aggregate by id, returning domain.ErrNotFound when absent.
func (r *Repository) FindByID(ctx context.Context, id domain.ReportId) (*domain.Report, error) {
	var row ReportRow
	err := r.db.WithContext(ctx).Where("id = ?", id.String()).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	sections, err := r.loadSections(ctx, row.ID)
	if err != nil {
		return nil, err
	}
	return reportFromRow(&row, sections)
}

// List returns reports matching the query plus the total count.
func (r *Repository) List(ctx context.Context, q domain.ListQuery) ([]*domain.Report, int64, error) {
	scope := r.db.WithContext(ctx).Model(&ReportRow{})
	if q.Status != "" {
		scope = scope.Where("status = ?", string(q.Status))
	}

	var total int64
	if err := scope.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	tx := scope.Order("uploaded_at desc, id asc")
	if q.Offset > 0 {
		tx = tx.Offset(q.Offset)
	}
	if q.Limit > 0 {
		tx = tx.Limit(q.Limit)
	}

	var rows []ReportRow
	if err := tx.Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	items := make([]*domain.Report, 0, len(rows))
	for i := range rows {
		sections, err := r.loadSections(ctx, rows[i].ID)
		if err != nil {
			return nil, 0, err
		}
		agg, err := reportFromRow(&rows[i], sections)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, agg)
	}
	return items, total, nil
}

// Delete removes the aggregate and (via cascade) its sections.
func (r *Repository) Delete(ctx context.Context, id domain.ReportId) error {
	return r.db.WithContext(ctx).Where("id = ?", id.String()).Delete(&ReportRow{}).Error
}

func (r *Repository) loadSections(ctx context.Context, reportID string) ([]*domain.Section, error) {
	var rows []SectionRow
	if err := r.db.WithContext(ctx).
		Where("report_id = ?", reportID).
		Order("sort_order asc, id asc").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]*domain.Section, 0, len(rows))
	for i := range rows {
		s, err := sectionFromRow(&rows[i])
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, nil
}
