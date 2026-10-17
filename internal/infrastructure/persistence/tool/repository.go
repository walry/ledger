package tool

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	domain "github.com/walry/ledger/internal/domain/tool"
)

// Repository is the GORM-backed implementation of tool.Repository.
type Repository struct {
	db *gorm.DB
}

// NewRepository constructs the repository.
func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

var _ domain.Repository = (*Repository)(nil)

// Save upserts the tool aggregate root.
func (r *Repository) Save(ctx context.Context, t *domain.Tool) error {
	if t == nil {
		return errors.New("tool: nil aggregate")
	}
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{UpdateAll: true}).
		Create(toolToRow(t)).Error
}

// FindByID loads a tool, returning domain.ErrNotFound when absent.
func (r *Repository) FindByID(ctx context.Context, id domain.ToolId) (*domain.Tool, error) {
	var row ToolRow
	err := r.db.WithContext(ctx).Where("id = ?", id.String()).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return toolFromRow(&row), nil
}

// FindByName loads a tool by its unique name, returning domain.ErrNotFound when
// absent.
func (r *Repository) FindByName(ctx context.Context, name string) (*domain.Tool, error) {
	var row ToolRow
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return toolFromRow(&row), nil
}

// FindByIDs batch-loads tools; missing ids are silently skipped.
func (r *Repository) FindByIDs(ctx context.Context, ids []domain.ToolId) ([]*domain.Tool, error) {
	if len(ids) == 0 {
		return make([]*domain.Tool, 0), nil
	}
	raw := make([]string, 0, len(ids))
	for _, id := range ids {
		raw = append(raw, id.String())
	}
	var rows []ToolRow
	if err := r.db.WithContext(ctx).Where("id IN ?", raw).Find(&rows).Error; err != nil {
		return nil, err
	}
	items := make([]*domain.Tool, 0, len(rows))
	for i := range rows {
		items = append(items, toolFromRow(&rows[i]))
	}
	return items, nil
}

// List returns tools plus the total count.
func (r *Repository) List(ctx context.Context, offset, limit int) ([]*domain.Tool, int64, error) {
	scope := r.db.WithContext(ctx).Model(&ToolRow{})

	var total int64
	if err := scope.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	tx := scope.Order("created_at asc, id asc")
	if offset > 0 {
		tx = tx.Offset(offset)
	}
	if limit > 0 {
		tx = tx.Limit(limit)
	}

	var rows []ToolRow
	if err := tx.Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	items := make([]*domain.Tool, 0, len(rows))
	for i := range rows {
		items = append(items, toolFromRow(&rows[i]))
	}
	return items, total, nil
}

// Delete removes a tool by id.
func (r *Repository) Delete(ctx context.Context, id domain.ToolId) error {
	return r.db.WithContext(ctx).Where("id = ?", id.String()).Delete(&ToolRow{}).Error
}
