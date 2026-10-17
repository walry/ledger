package service

import (
	"context"

	"github.com/walry/ledger/internal/domain/knowledge"
)

// RetrievedChunk 是检索命中的切片（值对象）。
type RetrievedChunk struct {
	KnowledgeBaseID knowledge.KnowledgeBaseId
	DocumentID      knowledge.DocumentId
	ChunkID         knowledge.ChunkId
	Content         string
	Score           float64
	Metadata        map[string]string
}

// Retriever 是向量检索端口（实现在 infrastructure/vectorstore）。
//
// 同时承担索引写入与检索读取，使知识库聚合与向量库之间的协作
// 通过该端口完成。
type Retriever interface {
	// Upsert 将文档的全部切片写入向量库（按文档覆盖）。
	Upsert(ctx context.Context, kbID knowledge.KnowledgeBaseId, doc *knowledge.Document) error
	// DeleteDocument 删除文档在向量库中的全部切片。
	DeleteDocument(ctx context.Context, kbID knowledge.KnowledgeBaseId, docID knowledge.DocumentId) error
	// Retrieve 在指定知识库中检索最相关的切片。
	Retrieve(ctx context.Context, query string, kbIDs []knowledge.KnowledgeBaseId, topK int) ([]RetrievedChunk, error)
}

// Reranker 是重排端口（可选）。
type Reranker interface {
	Rerank(ctx context.Context, query string, chunks []RetrievedChunk, topN int) ([]RetrievedChunk, error)
}
