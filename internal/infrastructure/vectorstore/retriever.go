package vectorstore

import (
	"context"

	"github.com/walry/ledger/internal/domain/knowledge/service"
)

// PassthroughReranker 不重新打分，仅按检索顺序截断到 topN。
type PassthroughReranker struct{}

var _ service.Reranker = (*PassthroughReranker)(nil)

// NewPassthroughReranker 构造直通重排器。
func NewPassthroughReranker() *PassthroughReranker { return &PassthroughReranker{} }

// Rerank 保留原顺序并截断到 topN。
func (PassthroughReranker) Rerank(_ context.Context, _ string, chunks []service.RetrievedChunk, topN int) ([]service.RetrievedChunk, error) {
	if topN <= 0 || topN >= len(chunks) {
		return chunks, nil
	}
	return chunks[:topN], nil
}
