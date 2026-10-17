package knowledge

import (
	"context"
	"fmt"

	"github.com/walry/ledger/internal/domain/ddd"
	knowledgedomain "github.com/walry/ledger/internal/domain/knowledge"
	"github.com/walry/ledger/pkg/clock"
	"github.com/walry/ledger/pkg/idgen"
)

// CreateKnowledgeBaseUseCase 创建知识库。
type CreateKnowledgeBaseUseCase struct {
	repo knowledgedomain.KnowledgeBaseRepository
	ids  idgen.Generator
	clk  clock.Clock
	bus  ddd.EventBus
}

// NewCreateKnowledgeBaseUseCase 创建知识库用例。
func NewCreateKnowledgeBaseUseCase(repo knowledgedomain.KnowledgeBaseRepository, ids idgen.Generator, clk clock.Clock, bus ddd.EventBus) *CreateKnowledgeBaseUseCase {
	return &CreateKnowledgeBaseUseCase{repo: repo, ids: ids, clk: clk, bus: bus}
}

// Execute 执行创建。
func (uc *CreateKnowledgeBaseUseCase) Execute(ctx context.Context, cmd CreateKnowledgeBaseCommand) (CreateKnowledgeBaseResult, error) {
	kb, err := knowledgedomain.NewKnowledgeBase(
		knowledgedomain.KnowledgeBaseId(uc.ids.New()),
		cmd.Name,
		cmd.Description,
		cmd.EmbeddingModel,
		cmd.Dimension,
		uc.clk.Now(),
	)
	if err != nil {
		return CreateKnowledgeBaseResult{}, fmt.Errorf("knowledge: create knowledge base: %w", err)
	}
	if err := uc.repo.Save(ctx, kb); err != nil {
		return CreateKnowledgeBaseResult{}, fmt.Errorf("knowledge: save knowledge base: %w", err)
	}
	if err := publishEvents(ctx, uc.bus, kb.PullEvents()); err != nil {
		return CreateKnowledgeBaseResult{}, fmt.Errorf("knowledge: publish events: %w", err)
	}
	return CreateKnowledgeBaseResult{KnowledgeBaseID: kb.ID().String()}, nil
}
