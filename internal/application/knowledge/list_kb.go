package knowledge

import (
	"context"
	"fmt"

	knowledgedomain "github.com/walry/ledger/internal/domain/knowledge"
)

// ListKnowledgeBasesUseCase 分页查询知识库。
type ListKnowledgeBasesUseCase struct {
	repo knowledgedomain.KnowledgeBaseRepository
}

// NewListKnowledgeBasesUseCase 创建知识库列表用例。
func NewListKnowledgeBasesUseCase(repo knowledgedomain.KnowledgeBaseRepository) *ListKnowledgeBasesUseCase {
	return &ListKnowledgeBasesUseCase{repo: repo}
}

// Execute 执行分页查询。
func (uc *ListKnowledgeBasesUseCase) Execute(ctx context.Context, cmd ListKnowledgeBasesCommand) (ListKnowledgeBasesResult, error) {
	bases, total, err := uc.repo.List(ctx, cmd.Offset, cmd.Limit)
	if err != nil {
		return ListKnowledgeBasesResult{}, fmt.Errorf("knowledge: list knowledge bases: %w", err)
	}
	items := make([]KnowledgeBaseSummaryDTO, 0, len(bases))
	for _, kb := range bases {
		items = append(items, KnowledgeBaseSummaryDTO{
			KnowledgeBaseID: kb.ID().String(),
			Name:            kb.Name(),
			Description:     kb.Description(),
			EmbeddingModel:  kb.EmbeddingModel(),
			Dimension:       kb.Dimension(),
			DocumentCount:   len(kb.Documents()),
			CreatedAt:       kb.CreatedAt(),
		})
	}
	return ListKnowledgeBasesResult{KnowledgeBases: items, Total: total}, nil
}
