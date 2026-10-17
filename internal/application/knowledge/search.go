package knowledge

import (
	"context"
	"fmt"

	knowledgedomain "github.com/walry/ledger/internal/domain/knowledge"
	knowledgesvc "github.com/walry/ledger/internal/domain/knowledge/service"
)

// SearchKnowledgeUseCase 在知识库中检索相关切片。
type SearchKnowledgeUseCase struct {
	retriever knowledgesvc.Retriever
}

// NewSearchKnowledgeUseCase 创建知识检索用例。
func NewSearchKnowledgeUseCase(retriever knowledgesvc.Retriever) *SearchKnowledgeUseCase {
	return &SearchKnowledgeUseCase{retriever: retriever}
}

// Execute 执行检索。
func (uc *SearchKnowledgeUseCase) Execute(ctx context.Context, q SearchQuery) (SearchResult, error) {
	kbIDs := make([]knowledgedomain.KnowledgeBaseId, 0, len(q.KnowledgeBaseIDs))
	for _, id := range q.KnowledgeBaseIDs {
		kbIDs = append(kbIDs, knowledgedomain.NewKnowledgeBaseId(id))
	}
	chunks, err := uc.retriever.Retrieve(ctx, q.Query, kbIDs, q.TopK)
	if err != nil {
		return SearchResult{}, fmt.Errorf("knowledge: retrieve: %w", err)
	}
	items := make([]ChunkDTO, 0, len(chunks))
	for _, c := range chunks {
		items = append(items, ChunkDTO{
			KnowledgeBaseID: c.KnowledgeBaseID.String(),
			DocumentID:      c.DocumentID.String(),
			ChunkID:         c.ChunkID.String(),
			Content:         c.Content,
			Score:           c.Score,
			Metadata:        c.Metadata,
		})
	}
	return SearchResult{Chunks: items}, nil
}
