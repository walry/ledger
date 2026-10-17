package conversation

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/walry/ledger/internal/domain/agent"
	domain "github.com/walry/ledger/internal/domain/conversation"
)

// Repository is the GORM-backed implementation of conversation.Repository.
type Repository struct {
	db *gorm.DB
}

// NewRepository constructs the repository.
func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

var _ domain.Repository = (*Repository)(nil)

// Save upserts the conversation root and replaces its turns and messages.
func (r *Repository) Save(ctx context.Context, c *domain.Conversation) error {
	if c == nil {
		return errors.New("conversation: nil aggregate")
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		root := conversationToRow(c)
		if err := tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(root).Error; err != nil {
			return err
		}
		// Deleting turns cascades to their messages.
		if err := tx.Where("conversation_id = ?", root.ID).Delete(&TurnRow{}).Error; err != nil {
			return err
		}

		turns := c.Turns()
		if len(turns) == 0 {
			return nil
		}
		turnRows := make([]TurnRow, 0, len(turns))
		messageRows := make([]MessageRow, 0)
		for _, t := range turns {
			turnRows = append(turnRows, *turnToRow(root.ID, t))
			for i, m := range t.Messages() {
				messageRows = append(messageRows, *messageToRow(t.ID().String(), i, m))
			}
		}
		if err := tx.Create(&turnRows).Error; err != nil {
			return err
		}
		if len(messageRows) == 0 {
			return nil
		}
		return tx.Create(&messageRows).Error
	})
}

// FindByID loads a conversation with its turns and messages.
func (r *Repository) FindByID(ctx context.Context, id domain.ConversationId) (*domain.Conversation, error) {
	var row ConversationRow
	err := r.db.WithContext(ctx).Where("id = ?", id.String()).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	turns, err := r.loadTurns(ctx, row.ID)
	if err != nil {
		return nil, err
	}
	return conversationFromRow(&row, turns), nil
}

// ListByAgent returns conversations of an agent plus the total count.
func (r *Repository) ListByAgent(ctx context.Context, agentID agent.AgentId, offset, limit int) ([]*domain.Conversation, int64, error) {
	scope := r.db.WithContext(ctx).Model(&ConversationRow{}).Where("agent_id = ?", agentID.String())
	return r.list(ctx, scope, offset, limit)
}

// List returns all conversations plus the total count.
func (r *Repository) List(ctx context.Context, offset, limit int) ([]*domain.Conversation, int64, error) {
	scope := r.db.WithContext(ctx).Model(&ConversationRow{})
	return r.list(ctx, scope, offset, limit)
}

// Delete removes a conversation and (via cascade) its turns and messages.
func (r *Repository) Delete(ctx context.Context, id domain.ConversationId) error {
	return r.db.WithContext(ctx).Where("id = ?", id.String()).Delete(&ConversationRow{}).Error
}

func (r *Repository) list(ctx context.Context, scope *gorm.DB, offset, limit int) ([]*domain.Conversation, int64, error) {
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

	var rows []ConversationRow
	if err := tx.Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	items := make([]*domain.Conversation, 0, len(rows))
	for i := range rows {
		turns, err := r.loadTurns(ctx, rows[i].ID)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, conversationFromRow(&rows[i], turns))
	}
	return items, total, nil
}

func (r *Repository) loadTurns(ctx context.Context, conversationID string) ([]*domain.Turn, error) {
	var turns []TurnRow
	if err := r.db.WithContext(ctx).
		Where("conversation_id = ?", conversationID).
		Order("idx asc, created_at asc").
		Find(&turns).Error; err != nil {
		return nil, err
	}
	if len(turns) == 0 {
		return make([]*domain.Turn, 0), nil
	}

	ids := make([]string, 0, len(turns))
	for i := range turns {
		ids = append(ids, turns[i].ID)
	}
	var messages []MessageRow
	if err := r.db.WithContext(ctx).
		Where("turn_id IN ?", ids).
		Order("turn_id asc, idx asc, id asc").
		Find(&messages).Error; err != nil {
		return nil, err
	}
	grouped := make(map[string][]domain.Message, len(turns))
	for i := range messages {
		grouped[messages[i].TurnID] = append(grouped[messages[i].TurnID], messageFromRow(&messages[i]))
	}

	out := make([]*domain.Turn, 0, len(turns))
	for i := range turns {
		msgs := grouped[turns[i].ID]
		if msgs == nil {
			msgs = make([]domain.Message, 0)
		}
		out = append(out, turnFromRow(&turns[i], msgs))
	}
	return out, nil
}
