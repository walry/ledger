package knowledge

import (
	"context"
	"errors"
	"fmt"

	knowledgedomain "github.com/walry/ledger/internal/domain/knowledge"
)

// GetKnowledgeBaseUseCase 查询单个知识库详情（含文档列表）。
type GetKnowledgeBaseUseCase struct {
	repo knowledgedomain.KnowledgeBaseRepository
	docs knowledgedomain.DocumentRepository
}

// NewGetKnowledgeBaseUseCase 创建知识库详情用例。
func NewGetKnowledgeBaseUseCase(repo knowledgedomain.KnowledgeBaseRepository, docs knowledgedomain.DocumentRepository) *GetKnowledgeBaseUseCase {
	return &GetKnowledgeBaseUseCase{repo: repo, docs: docs}
}

// Execute 执行查询。
func (uc *GetKnowledgeBaseUseCase) Execute(ctx context.Context, cmd GetKnowledgeBaseCommand) (KnowledgeBaseDetailDTO, error) {
	kb, err := uc.repo.FindByID(ctx, knowledgedomain.NewKnowledgeBaseId(cmd.KnowledgeBaseID))
	if err != nil {
		if errors.Is(err, knowledgedomain.ErrNotFound) {
			return KnowledgeBaseDetailDTO{}, fmt.Errorf("knowledge: %s not found: %w", cmd.KnowledgeBaseID, err)
		}
		return KnowledgeBaseDetailDTO{}, fmt.Errorf("knowledge: load knowledge base: %w", err)
	}

	docs, err := uc.docs.ListByKnowledgeBase(ctx, kb.ID())
	if err != nil {
		return KnowledgeBaseDetailDTO{}, fmt.Errorf("knowledge: list documents: %w", err)
	}
	docDTOs := make([]DocumentDTO, 0, len(docs))
	for _, d := range docs {
		src := d.Source()
		docDTOs = append(docDTOs, DocumentDTO{
			DocumentID:    d.ID().String(),
			Title:         d.Title(),
			SourceKind:    string(src.Kind),
			SourceRefID:   src.RefID,
			Status:        string(d.Status()),
			FailureReason: d.FailureReason(),
			ChunkCount:    len(d.Chunks()),
			CreatedAt:     d.CreatedAt(),
		})
	}

	return KnowledgeBaseDetailDTO{
		KnowledgeBaseID: kb.ID().String(),
		Name:            kb.Name(),
		Description:     kb.Description(),
		EmbeddingModel:  kb.EmbeddingModel(),
		Dimension:       kb.Dimension(),
		DocumentCount:   len(docs),
		CreatedAt:       kb.CreatedAt(),
		UpdatedAt:       kb.UpdatedAt(),
		Documents:       docDTOs,
	}, nil
}
