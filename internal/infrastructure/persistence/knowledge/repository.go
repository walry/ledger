package knowledge

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	domain "github.com/walry/ledger/internal/domain/knowledge"
)

// Repository is the GORM-backed store shared by the KnowledgeBase and Document
// aggregates.
//
// Note: the two domain ports (KnowledgeBaseRepository, DocumentRepository)
// declare identically-named methods (Save/FindByID/Delete) with *different*
// parameter types. Go has no method overloading, so no single concrete type can
// satisfy both interfaces literally. Repository therefore implements
// KnowledgeBaseRepository directly and exposes the document side through
// Documents(), which returns a DocumentRepository-declared adapter sharing the
// same *gorm.DB.
type Repository struct {
	db *gorm.DB
}

// NewRepository constructs the repository.
func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

// Documents returns an adapter implementing the Document aggregate port.
func (r *Repository) Documents() *DocumentRepository { return &DocumentRepository{db: r.db} }

var _ domain.KnowledgeBaseRepository = (*Repository)(nil)

// --- KnowledgeBase aggregate ---

// Save upserts the knowledge base aggregate root.
func (r *Repository) Save(ctx context.Context, kb *domain.KnowledgeBase) error {
	if kb == nil {
		return errors.New("knowledge: nil knowledge base")
	}
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{UpdateAll: true}).
		Create(knowledgeBaseToRow(kb)).Error
}

// FindByID loads a knowledge base, returning domain.ErrNotFound when absent.
func (r *Repository) FindByID(ctx context.Context, id domain.KnowledgeBaseId) (*domain.KnowledgeBase, error) {
	var row KnowledgeBaseRow
	err := r.db.WithContext(ctx).Where("id = ?", id.String()).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return knowledgeBaseFromRow(&row), nil
}

// List returns knowledge bases plus the total count.
func (r *Repository) List(ctx context.Context, offset, limit int) ([]*domain.KnowledgeBase, int64, error) {
	scope := r.db.WithContext(ctx).Model(&KnowledgeBaseRow{})

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

	var rows []KnowledgeBaseRow
	if err := tx.Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	items := make([]*domain.KnowledgeBase, 0, len(rows))
	for i := range rows {
		items = append(items, knowledgeBaseFromRow(&rows[i]))
	}
	return items, total, nil
}

// Delete removes a knowledge base by id.
func (r *Repository) Delete(ctx context.Context, id domain.KnowledgeBaseId) error {
	return r.db.WithContext(ctx).Where("id = ?", id.String()).Delete(&KnowledgeBaseRow{}).Error
}

// DocumentRepository is the adapter implementing domain.DocumentRepository.
type DocumentRepository struct {
	db *gorm.DB
}

var _ domain.DocumentRepository = (*DocumentRepository)(nil)

// Save upserts the document root and replaces its chunks inside one tx.
func (r *DocumentRepository) Save(ctx context.Context, d *domain.Document) error {
	if d == nil {
		return errors.New("knowledge: nil document")
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		root := documentToRow(d)
		if err := tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(root).Error; err != nil {
			return err
		}
		if err := tx.Where("document_id = ?", root.ID).Delete(&ChunkRow{}).Error; err != nil {
			return err
		}
		chunks := d.Chunks()
		if len(chunks) == 0 {
			return nil
		}
		rows := make([]ChunkRow, 0, len(chunks))
		for _, c := range chunks {
			rows = append(rows, *chunkToRow(root.ID, c))
		}
		return tx.Create(&rows).Error
	})
}

// FindByID loads a document with its chunks.
func (r *DocumentRepository) FindByID(ctx context.Context, id domain.DocumentId) (*domain.Document, error) {
	var row DocumentRow
	err := r.db.WithContext(ctx).Where("id = ?", id.String()).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	chunks, err := r.loadChunks(ctx, row.ID)
	if err != nil {
		return nil, err
	}
	return documentFromRow(&row, chunks), nil
}

// FindBySource finds a document by its source reference (event idempotency).
func (r *DocumentRepository) FindBySource(ctx context.Context, kbID domain.KnowledgeBaseId, kind domain.SourceKind, refID string) (*domain.Document, error) {
	var row DocumentRow
	err := r.db.WithContext(ctx).
		Where("kb_id = ? AND source_kind = ? AND source_ref_id = ?", kbID.String(), string(kind), refID).
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	chunks, err := r.loadChunks(ctx, row.ID)
	if err != nil {
		return nil, err
	}
	return documentFromRow(&row, chunks), nil
}

// ListByKnowledgeBase returns all documents of a knowledge base, with chunks.
func (r *DocumentRepository) ListByKnowledgeBase(ctx context.Context, kbID domain.KnowledgeBaseId) ([]*domain.Document, error) {
	var rows []DocumentRow
	if err := r.db.WithContext(ctx).
		Where("kb_id = ?", kbID.String()).
		Order("created_at asc, id asc").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return make([]*domain.Document, 0), nil
	}

	ids := make([]string, 0, len(rows))
	for i := range rows {
		ids = append(ids, rows[i].ID)
	}
	var chunkRows []ChunkRow
	if err := r.db.WithContext(ctx).
		Where("document_id IN ?", ids).
		Order("document_id asc, idx asc").
		Find(&chunkRows).Error; err != nil {
		return nil, err
	}
	grouped := make(map[string][]*domain.Chunk, len(rows))
	for i := range chunkRows {
		grouped[chunkRows[i].DocumentID] = append(grouped[chunkRows[i].DocumentID], chunkFromRow(&chunkRows[i]))
	}

	items := make([]*domain.Document, 0, len(rows))
	for i := range rows {
		chunks := grouped[rows[i].ID]
		if chunks == nil {
			chunks = make([]*domain.Chunk, 0)
		}
		items = append(items, documentFromRow(&rows[i], chunks))
	}
	return items, nil
}

// Delete removes a document and (via cascade) its chunks.
func (r *DocumentRepository) Delete(ctx context.Context, id domain.DocumentId) error {
	return r.db.WithContext(ctx).Where("id = ?", id.String()).Delete(&DocumentRow{}).Error
}

func (r *DocumentRepository) loadChunks(ctx context.Context, documentID string) ([]*domain.Chunk, error) {
	var rows []ChunkRow
	if err := r.db.WithContext(ctx).
		Where("document_id = ?", documentID).
		Order("idx asc, id asc").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]*domain.Chunk, 0, len(rows))
	for i := range rows {
		out = append(out, chunkFromRow(&rows[i]))
	}
	return out, nil
}
