package vectorstore

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/pgvector/pgvector-go"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/walry/ledger/internal/domain/knowledge"
	"github.com/walry/ledger/internal/domain/knowledge/service"
)

const tableName = "knowledge_vectors"

// defaultTopK 是 topK <= 0 时的默认检索条数。
const defaultTopK = 10

// PgVectorStore 基于 PostgreSQL + pgvector 实现知识库向量检索。
type PgVectorStore struct {
	db        *gorm.DB
	embedder  service.Embedder
	dimension int
}

// NewPgVectorStore 构造 pgvector 向量库适配器。
func NewPgVectorStore(db *gorm.DB, embedder service.Embedder, dimension int) *PgVectorStore {
	return &PgVectorStore{db: db, embedder: embedder, dimension: dimension}
}

var _ service.Retriever = (*PgVectorStore)(nil)

// vectorRow 映射 knowledge_vectors 表的一行。
type vectorRow struct {
	ID         int64           `gorm:"column:id;primaryKey"`
	KBID       string          `gorm:"column:kb_id"`
	DocumentID string          `gorm:"column:document_id"`
	ChunkID    string          `gorm:"column:chunk_id"`
	Title      string          `gorm:"column:title"`
	Content    string          `gorm:"column:content"`
	Metadata   string          `gorm:"column:metadata;type:jsonb"`
	Embedding  pgvector.Vector `gorm:"column:embedding;type:vector"`
}

func (vectorRow) TableName() string { return tableName }

// retrievalHit 是检索查询的扫描目标。
type retrievalHit struct {
	KBID       string  `gorm:"column:kb_id"`
	DocumentID string  `gorm:"column:document_id"`
	ChunkID    string  `gorm:"column:chunk_id"`
	Content    string  `gorm:"column:content"`
	Metadata   string  `gorm:"column:metadata"`
	Distance   float64 `gorm:"column:distance"`
}

// EnsureSchema 创建 vector 扩展、向量表与索引（幂等）。
func (s *PgVectorStore) EnsureSchema(ctx context.Context) error {
	if err := s.db.WithContext(ctx).Exec("CREATE EXTENSION IF NOT EXISTS vector").Error; err != nil {
		return fmt.Errorf("vectorstore: create extension: %w", err)
	}

	createTable := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
		id bigserial PRIMARY KEY,
		kb_id text NOT NULL,
		document_id text NOT NULL,
		chunk_id text NOT NULL,
		title text NOT NULL DEFAULT '',
		content text NOT NULL DEFAULT '',
		metadata jsonb,
		embedding vector(%d) NOT NULL
	)`, tableName, s.dimension)
	if err := s.db.WithContext(ctx).Exec(createTable).Error; err != nil {
		return fmt.Errorf("vectorstore: create table: %w", err)
	}

	statements := []string{
		fmt.Sprintf("CREATE INDEX IF NOT EXISTS %s_embedding_idx ON %s USING hnsw (embedding vector_cosine_ops)", tableName, tableName),
		fmt.Sprintf("CREATE INDEX IF NOT EXISTS %s_kb_document_idx ON %s (kb_id, document_id)", tableName, tableName),
	}
	for _, stmt := range statements {
		if err := s.db.WithContext(ctx).Exec(stmt).Error; err != nil {
			return fmt.Errorf("vectorstore: create index: %w", err)
		}
	}
	return nil
}

// Upsert 按文档覆盖写入向量库：先删旧行，再写入全部切片。
func (s *PgVectorStore) Upsert(ctx context.Context, kbID knowledge.KnowledgeBaseId, doc *knowledge.Document) error {
	if doc == nil {
		return fmt.Errorf("vectorstore: document is nil")
	}

	rows, err := s.buildRows(ctx, kbID, doc)
	if err != nil {
		return err
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("DELETE FROM "+tableName+" WHERE kb_id = ? AND document_id = ?",
			kbID.String(), doc.ID().String()).Error; err != nil {
			return fmt.Errorf("vectorstore: delete old rows: %w", err)
		}
		if len(rows) == 0 {
			return nil
		}
		if err := tx.Create(&rows).Error; err != nil {
			return fmt.Errorf("vectorstore: insert rows: %w", err)
		}
		return nil
	})
}

// DeleteDocument 删除文档在向量库中的全部切片。
func (s *PgVectorStore) DeleteDocument(ctx context.Context, kbID knowledge.KnowledgeBaseId, docID knowledge.DocumentId) error {
	if err := s.db.WithContext(ctx).
		Exec("DELETE FROM "+tableName+" WHERE kb_id = ? AND document_id = ?", kbID.String(), docID.String()).Error; err != nil {
		return fmt.Errorf("vectorstore: delete document: %w", err)
	}
	return nil
}

// Retrieve 在指定知识库中检索最相关的切片。
func (s *PgVectorStore) Retrieve(ctx context.Context, query string, kbIDs []knowledge.KnowledgeBaseId, topK int) ([]service.RetrievedChunk, error) {
	if len(kbIDs) == 0 {
		return nil, nil
	}
	if topK <= 0 {
		topK = defaultTopK
	}
	if s.embedder == nil {
		return nil, fmt.Errorf("vectorstore: embedder is required for retrieval")
	}

	embs, err := s.embedder.Embed(ctx, []string{query})
	if err != nil {
		return nil, fmt.Errorf("vectorstore: embed query: %w", err)
	}
	if len(embs) == 0 {
		return nil, fmt.Errorf("vectorstore: embedder returned no vector for query")
	}
	vec := pgvector.NewVector(embs[0].Vector())

	ids := make([]string, len(kbIDs))
	for i, id := range kbIDs {
		ids[i] = id.String()
	}

	var hits []retrievalHit
	err = s.db.WithContext(ctx).
		Table(tableName).
		Select("kb_id, document_id, chunk_id, content, metadata, embedding <=> ? AS distance", vec).
		Where("kb_id = ANY(?)", ids).
		Clauses(clause.OrderBy{Expression: gorm.Expr("embedding <=> ?", vec)}).
		Limit(topK).
		Scan(&hits).Error
	if err != nil {
		return nil, fmt.Errorf("vectorstore: query: %w", err)
	}

	out := make([]service.RetrievedChunk, 0, len(hits))
	for _, h := range hits {
		out = append(out, service.RetrievedChunk{
			KnowledgeBaseID: knowledge.NewKnowledgeBaseId(h.KBID),
			DocumentID:      knowledge.NewDocumentId(h.DocumentID),
			ChunkID:         knowledge.NewChunkId(h.ChunkID),
			Content:         h.Content,
			Score:           clampScore(1 - h.Distance),
			Metadata:        unmarshalMetadata(h.Metadata),
		})
	}
	return out, nil
}

// buildRows 将文档切片转为待写入行；缺失向量的切片用 embedder 补齐。
func (s *PgVectorStore) buildRows(ctx context.Context, kbID knowledge.KnowledgeBaseId, doc *knowledge.Document) ([]vectorRow, error) {
	chunks := doc.Chunks()
	rows := make([]vectorRow, len(chunks))

	var pendingTexts []string
	var pendingIdx []int
	for i, chunk := range chunks {
		meta, err := marshalMetadata(chunk.Metadata())
		if err != nil {
			return nil, fmt.Errorf("vectorstore: encode metadata: %w", err)
		}
		rows[i] = vectorRow{
			KBID:       kbID.String(),
			DocumentID: doc.ID().String(),
			ChunkID:    chunk.ID().String(),
			Title:      doc.Title(),
			Content:    chunk.Content(),
			Metadata:   meta,
		}
		if emb := chunk.Embedding(); !emb.IsZero() {
			rows[i].Embedding = pgvector.NewVector(emb.Vector())
			continue
		}
		pendingTexts = append(pendingTexts, chunk.Content())
		pendingIdx = append(pendingIdx, i)
	}

	if len(pendingTexts) > 0 {
		if s.embedder == nil {
			return nil, fmt.Errorf("vectorstore: embedder is required for chunks without embeddings")
		}
		embs, err := s.embedder.Embed(ctx, pendingTexts)
		if err != nil {
			return nil, fmt.Errorf("vectorstore: embed chunks: %w", err)
		}
		if len(embs) != len(pendingTexts) {
			return nil, fmt.Errorf("vectorstore: embedder returned %d vectors for %d texts", len(embs), len(pendingTexts))
		}
		for j, idx := range pendingIdx {
			rows[idx].Embedding = pgvector.NewVector(embs[j].Vector())
		}
	}
	return rows, nil
}

func marshalMetadata(m map[string]string) (string, error) {
	if len(m) == 0 {
		return "{}", nil
	}
	b, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func unmarshalMetadata(s string) map[string]string {
	if s == "" || s == "null" {
		return nil
	}
	var m map[string]string
	if err := json.Unmarshal([]byte(s), &m); err != nil {
		return nil
	}
	return m
}

// clampScore 将余弦相似度限制到 [-1, 1]。
func clampScore(v float64) float64 {
	switch {
	case v > 1:
		return 1
	case v < -1:
		return -1
	default:
		return v
	}
}
