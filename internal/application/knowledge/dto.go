package knowledge

import (
	"context"
	"time"

	"github.com/walry/ledger/internal/domain/ddd"
	knowledgedomain "github.com/walry/ledger/internal/domain/knowledge"
)

// CreateKnowledgeBaseCommand 是创建知识库的输入。
type CreateKnowledgeBaseCommand struct {
	Name           string
	Description    string
	EmbeddingModel string
	Dimension      int
}

// CreateKnowledgeBaseResult 是创建知识库的输出。
type CreateKnowledgeBaseResult struct {
	KnowledgeBaseID string
}

// AddDocumentCommand 是新增文档的输入。
type AddDocumentCommand struct {
	KnowledgeBaseID string
	Title           string
	SourceKind      knowledgedomain.SourceKind
	SourceRefID     string
	Text            string
}

// AddDocumentResult 是新增文档的输出。
type AddDocumentResult struct {
	DocumentID string
	ChunkCount int
}

// ReindexDocumentCommand 是重建索引的输入。Text 为空时使用文档现有切片文本。
type ReindexDocumentCommand struct {
	DocumentID string
	Text       string
}

// ReindexDocumentResult 是重建索引的输出。
type ReindexDocumentResult struct {
	DocumentID string
	ChunkCount int
}

// SearchQuery 是检索知识库的输入。
type SearchQuery struct {
	Query            string
	KnowledgeBaseIDs []string
	TopK             int
}

// SearchResult 是检索知识库的输出。
type SearchResult struct {
	Chunks []ChunkDTO
}

// ChunkDTO 是检索命中的切片 DTO。
type ChunkDTO struct {
	KnowledgeBaseID string
	DocumentID      string
	ChunkID         string
	Content         string
	Score           float64
	Metadata        map[string]string
}

// ListKnowledgeBasesCommand 是分页查询知识库的输入。
type ListKnowledgeBasesCommand struct {
	Offset int
	Limit  int
}

// ListKnowledgeBasesResult 是知识库列表 DTO。
type ListKnowledgeBasesResult struct {
	KnowledgeBases []KnowledgeBaseSummaryDTO
	Total          int64
}

// KnowledgeBaseSummaryDTO 是知识库列表项 DTO。
type KnowledgeBaseSummaryDTO struct {
	KnowledgeBaseID string
	Name            string
	Description     string
	EmbeddingModel  string
	Dimension       int
	DocumentCount   int
	CreatedAt       time.Time
}

// GetKnowledgeBaseCommand 是查询单个知识库的输入。
type GetKnowledgeBaseCommand struct {
	KnowledgeBaseID string
}

// KnowledgeBaseDetailDTO 是知识库详情 DTO，含文档列表。
type KnowledgeBaseDetailDTO struct {
	KnowledgeBaseID string
	Name            string
	Description     string
	EmbeddingModel  string
	Dimension       int
	DocumentCount   int
	CreatedAt       time.Time
	UpdatedAt       time.Time
	Documents       []DocumentDTO
}

// DocumentDTO 是文档 DTO。
type DocumentDTO struct {
	DocumentID    string
	Title         string
	SourceKind    string
	SourceRefID   string
	Status        string
	FailureReason string
	ChunkCount    int
	CreatedAt     time.Time
}

// publishEvents 发布聚合累积的领域事件，bus 为空时安全跳过。
func publishEvents(ctx context.Context, bus ddd.EventBus, events []ddd.DomainEvent) error {
	if bus == nil || len(events) == 0 {
		return nil
	}
	return bus.Publish(ctx, events...)
}
