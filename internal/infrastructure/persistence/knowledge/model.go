// Package knowledge provides the GORM persistence adapter for the knowledge
// bounded context. It backs both the KnowledgeBase aggregate and the Document
// aggregate.
package knowledge

import (
	"encoding/json"
	"time"

	domain "github.com/walry/ledger/internal/domain/knowledge"
)

// KnowledgeBaseRow is the persistence model for the KnowledgeBase aggregate.
type KnowledgeBaseRow struct {
	ID             string    `gorm:"column:id;primaryKey;type:text"`
	Name           string    `gorm:"column:name;type:text;not null"`
	Description    string    `gorm:"column:description;type:text"`
	EmbeddingModel string    `gorm:"column:embedding_model;type:text"`
	Dimension      int       `gorm:"column:dimension;not null;default:0"`
	Documents      []byte    `gorm:"column:documents;type:jsonb"`
	CreatedAt      time.Time `gorm:"column:created_at;not null"`
	UpdatedAt      time.Time `gorm:"column:updated_at;not null"`
}

// TableName pins the physical table name.
func (KnowledgeBaseRow) TableName() string { return "knowledge_bases" }

// DocumentRow is the persistence model for the Document aggregate root.
type DocumentRow struct {
	ID            string    `gorm:"column:id;primaryKey;type:text"`
	KnowledgeBase string    `gorm:"column:kb_id;type:text;not null;index"`
	SourceKind    string    `gorm:"column:source_kind;type:text;not null"`
	SourceRefID   string    `gorm:"column:source_ref_id;type:text;not null"`
	SourceTitle   string    `gorm:"column:source_title;type:text"`
	Title         string    `gorm:"column:title;type:text"`
	Status        string    `gorm:"column:status;type:text;not null;index"`
	FailureReason string    `gorm:"column:failure_reason;type:text"`
	CreatedAt     time.Time `gorm:"column:created_at;not null"`
	UpdatedAt     time.Time `gorm:"column:updated_at;not null"`
}

// TableName pins the physical table name.
func (DocumentRow) TableName() string { return "knowledge_documents" }

// ChunkRow is the persistence model for the Chunk entity inside the Document
// aggregate. It is a child table with an ON DELETE CASCADE foreign key.
type ChunkRow struct {
	ID              string `gorm:"column:id;primaryKey;type:text"`
	DocumentID      string `gorm:"column:document_id;type:text;not null;index"`
	Index           int    `gorm:"column:idx;not null;default:0"`
	Content         string `gorm:"column:content;type:text"`
	EmbeddingVector []byte `gorm:"column:embedding_vector;type:jsonb"`
	EmbeddingModel  string `gorm:"column:embedding_model;type:text"`
	Metadata        []byte `gorm:"column:metadata;type:jsonb"`
	TokenCount      int    `gorm:"column:token_count;not null;default:0"`

	Document DocumentRow `gorm:"foreignKey:DocumentID;references:ID;constraint:OnDelete:CASCADE"`
}

// TableName pins the physical table name.
func (ChunkRow) TableName() string { return "knowledge_chunks" }

func mustJSON(v any, fallback string) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		return []byte(fallback)
	}
	return b
}

func documentsToJSON(refs []domain.DocumentRef) []byte {
	if refs == nil {
		refs = make([]domain.DocumentRef, 0)
	}
	return mustJSON(refs, "[]")
}

func documentsFromJSON(b []byte) []domain.DocumentRef {
	out := make([]domain.DocumentRef, 0)
	if len(b) == 0 {
		return out
	}
	if err := json.Unmarshal(b, &out); err != nil {
		return make([]domain.DocumentRef, 0)
	}
	if out == nil {
		out = make([]domain.DocumentRef, 0)
	}
	return out
}

func metadataToJSON(m map[string]string) []byte {
	if m == nil {
		m = map[string]string{}
	}
	return mustJSON(m, "{}")
}

func metadataFromJSON(b []byte) map[string]string {
	out := map[string]string{}
	if len(b) == 0 {
		return out
	}
	if err := json.Unmarshal(b, &out); err != nil {
		return map[string]string{}
	}
	if out == nil {
		out = map[string]string{}
	}
	return out
}

func embeddingToRow(e domain.Embedding) (vector []byte, model string) {
	return mustJSON(e.Vector(), "[]"), e.Model()
}

func embeddingFromRow(vector []byte, model string) domain.Embedding {
	if len(vector) == 0 {
		return domain.Embedding{}
	}
	var vec []float32
	if err := json.Unmarshal(vector, &vec); err != nil || len(vec) == 0 {
		return domain.Embedding{}
	}
	emb, err := domain.NewEmbedding(model, vec)
	if err != nil {
		return domain.Embedding{}
	}
	return emb
}

func knowledgeBaseToRow(k *domain.KnowledgeBase) *KnowledgeBaseRow {
	return &KnowledgeBaseRow{
		ID:             k.ID().String(),
		Name:           k.Name(),
		Description:    k.Description(),
		EmbeddingModel: k.EmbeddingModel(),
		Dimension:      k.Dimension(),
		Documents:      documentsToJSON(k.Documents()),
		CreatedAt:      k.CreatedAt(),
		UpdatedAt:      k.UpdatedAt(),
	}
}

func knowledgeBaseFromRow(row *KnowledgeBaseRow) *domain.KnowledgeBase {
	return domain.RehydrateKnowledgeBase(
		domain.KnowledgeBaseId(row.ID), row.Name, row.Description, row.EmbeddingModel,
		row.Dimension, documentsFromJSON(row.Documents), row.CreatedAt, row.UpdatedAt,
	)
}

func documentToRow(d *domain.Document) *DocumentRow {
	src := d.Source()
	return &DocumentRow{
		ID:            d.ID().String(),
		KnowledgeBase: d.KnowledgeBaseID().String(),
		SourceKind:    string(src.Kind),
		SourceRefID:   src.RefID,
		SourceTitle:   src.Title,
		Title:         d.Title(),
		Status:        string(d.Status()),
		FailureReason: d.FailureReason(),
		CreatedAt:     d.CreatedAt(),
		UpdatedAt:     d.UpdatedAt(),
	}
}

func documentFromRow(row *DocumentRow, chunks []*domain.Chunk) *domain.Document {
	src := domain.SourceRef{
		Kind:  domain.SourceKind(row.SourceKind),
		RefID: row.SourceRefID,
		Title: row.SourceTitle,
	}
	return domain.RehydrateDocument(
		domain.DocumentId(row.ID), domain.KnowledgeBaseId(row.KnowledgeBase), src, row.Title,
		chunks, domain.DocumentStatus(row.Status), row.FailureReason, row.CreatedAt, row.UpdatedAt,
	)
}

func chunkToRow(documentID string, c *domain.Chunk) *ChunkRow {
	vector, model := embeddingToRow(c.Embedding())
	return &ChunkRow{
		ID:              c.ID().String(),
		DocumentID:      documentID,
		Index:           c.Index(),
		Content:         c.Content(),
		EmbeddingVector: vector,
		EmbeddingModel:  model,
		Metadata:        metadataToJSON(c.Metadata()),
		TokenCount:      c.TokenCount(),
	}
}

func chunkFromRow(row *ChunkRow) *domain.Chunk {
	emb := embeddingFromRow(row.EmbeddingVector, row.EmbeddingModel)
	return domain.RehydrateChunk(
		domain.ChunkId(row.ID), row.Index, row.Content, emb, metadataFromJSON(row.Metadata), row.TokenCount,
	)
}
