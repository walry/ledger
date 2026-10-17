package knowledge

import (
	"context"
	"errors"
)

// ErrNotFound 表示目标不存在。
var ErrNotFound = errors.New("knowledge: not found")

// KnowledgeBaseRepository 是知识库聚合的仓储接口。
type KnowledgeBaseRepository interface {
	Save(ctx context.Context, kb *KnowledgeBase) error
	FindByID(ctx context.Context, id KnowledgeBaseId) (*KnowledgeBase, error)
	List(ctx context.Context, offset, limit int) ([]*KnowledgeBase, int64, error)
	Delete(ctx context.Context, id KnowledgeBaseId) error
}

// DocumentRepository 是文档聚合的仓储接口。
type DocumentRepository interface {
	Save(ctx context.Context, d *Document) error
	FindByID(ctx context.Context, id DocumentId) (*Document, error)
	// FindBySource 依据来源引用查找文档，用于事件处理幂等。
	FindBySource(ctx context.Context, kbID KnowledgeBaseId, kind SourceKind, refID string) (*Document, error)
	ListByKnowledgeBase(ctx context.Context, kbID KnowledgeBaseId) ([]*Document, error)
	Delete(ctx context.Context, id DocumentId) error
}
