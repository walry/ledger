package modelconfig

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	domain "github.com/walry/ledger/internal/domain/modelconfig"
)

// Repository is the GORM-backed implementation of modelconfig.Repository.
type Repository struct {
	db *gorm.DB
}

// NewRepository constructs the repository.
func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

var _ domain.Repository = (*Repository)(nil)

// Save upserts the model config aggregate root.
func (r *Repository) Save(ctx context.Context, m *domain.ModelConfig) error {
	if m == nil {
		return errors.New("modelconfig: nil aggregate")
	}
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{UpdateAll: true}).
		Create(modelConfigToRow(m)).Error
}

// FindByID loads a config, returning domain.ErrNotFound when absent.
func (r *Repository) FindByID(ctx context.Context, id domain.ModelConfigId) (*domain.ModelConfig, error) {
	var row ModelConfigRow
	err := r.db.WithContext(ctx).Where("id = ?", id.String()).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return modelConfigFromRow(&row), nil
}

// List returns configs plus the total count.
func (r *Repository) List(ctx context.Context, offset, limit int) ([]*domain.ModelConfig, int64, error) {
	scope := r.db.WithContext(ctx).Model(&ModelConfigRow{})

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

	var rows []ModelConfigRow
	if err := tx.Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	items := make([]*domain.ModelConfig, 0, len(rows))
	for i := range rows {
		items = append(items, modelConfigFromRow(&rows[i]))
	}
	return items, total, nil
}

// Delete removes a config by id.
func (r *Repository) Delete(ctx context.Context, id domain.ModelConfigId) error {
	return r.db.WithContext(ctx).Where("id = ?", id.String()).Delete(&ModelConfigRow{}).Error
}
