package agent

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	domain "github.com/walry/ledger/internal/domain/agent"
)

// Repository is the GORM-backed implementation of agent.Repository.
type Repository struct {
	db *gorm.DB
}

// NewRepository constructs the repository.
func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

var _ domain.Repository = (*Repository)(nil)

// Save upserts the agent aggregate root.
func (r *Repository) Save(ctx context.Context, a *domain.Agent) error {
	if a == nil {
		return errors.New("agent: nil aggregate")
	}
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{UpdateAll: true}).
		Create(agentToRow(a)).Error
}

// FindByID loads an agent, returning domain.ErrNotFound when absent.
func (r *Repository) FindByID(ctx context.Context, id domain.AgentId) (*domain.Agent, error) {
	var row AgentRow
	err := r.db.WithContext(ctx).Where("id = ?", id.String()).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return agentFromRow(&row), nil
}

// FindPublished returns the latest published version of the agent family that
// the given id belongs to. The family is identified by the agent name (new
// versions derive from a base agent and keep the same name, with version
// increasing). Returns domain.ErrNotFound when the id or any published version
// is missing.
func (r *Repository) FindPublished(ctx context.Context, id domain.AgentId) (*domain.Agent, error) {
	var seed AgentRow
	err := r.db.WithContext(ctx).
		Select("name").
		Where("id = ?", id.String()).
		First(&seed).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	var row AgentRow
	err = r.db.WithContext(ctx).
		Where("name = ? AND status = ?", seed.Name, string(domain.StatusPublished)).
		Order("version desc, updated_at desc").
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return agentFromRow(&row), nil
}

// List returns agents plus the total count.
func (r *Repository) List(ctx context.Context, offset, limit int) ([]*domain.Agent, int64, error) {
	scope := r.db.WithContext(ctx).Model(&AgentRow{})

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

	var rows []AgentRow
	if err := tx.Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	items := make([]*domain.Agent, 0, len(rows))
	for i := range rows {
		items = append(items, agentFromRow(&rows[i]))
	}
	return items, total, nil
}

// Delete removes an agent by id.
func (r *Repository) Delete(ctx context.Context, id domain.AgentId) error {
	return r.db.WithContext(ctx).Where("id = ?", id.String()).Delete(&AgentRow{}).Error
}
