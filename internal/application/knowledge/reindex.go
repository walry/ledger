package knowledge

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/walry/ledger/internal/domain/ddd"
	knowledgedomain "github.com/walry/ledger/internal/domain/knowledge"
	knowledgesvc "github.com/walry/ledger/internal/domain/knowledge/service"
	"github.com/walry/ledger/pkg/clock"
	"github.com/walry/ledger/pkg/idgen"
)

// ReindexDocumentUseCase 重建文档索引：重新切片、重新向量化并覆盖向量库。
type ReindexDocumentUseCase struct {
	kbs       knowledgedomain.KnowledgeBaseRepository
	docs      knowledgedomain.DocumentRepository
	chunker   knowledgesvc.Chunker
	embedder  knowledgesvc.Embedder
	retriever knowledgesvc.Retriever
	ids       idgen.Generator
	clk       clock.Clock
	bus       ddd.EventBus
}

// NewReindexDocumentUseCase 创建重建索引用例。
func NewReindexDocumentUseCase(
	kbs knowledgedomain.KnowledgeBaseRepository,
	docs knowledgedomain.DocumentRepository,
	chunker knowledgesvc.Chunker,
	embedder knowledgesvc.Embedder,
	retriever knowledgesvc.Retriever,
	ids idgen.Generator,
	clk clock.Clock,
	bus ddd.EventBus,
) *ReindexDocumentUseCase {
	return &ReindexDocumentUseCase{
		kbs: kbs, docs: docs, chunker: chunker, embedder: embedder,
		retriever: retriever, ids: ids, clk: clk, bus: bus,
	}
}

// Execute 执行重建索引。
func (uc *ReindexDocumentUseCase) Execute(ctx context.Context, cmd ReindexDocumentCommand) (ReindexDocumentResult, error) {
	docID := knowledgedomain.NewDocumentId(cmd.DocumentID)
	doc, err := uc.docs.FindByID(ctx, docID)
	if err != nil {
		if errors.Is(err, knowledgedomain.ErrNotFound) {
			return ReindexDocumentResult{}, fmt.Errorf("knowledge: document %s not found: %w", cmd.DocumentID, err)
		}
		return ReindexDocumentResult{}, fmt.Errorf("knowledge: load document: %w", err)
	}
	kbID := doc.KnowledgeBaseID()

	text := strings.TrimSpace(cmd.Text)
	if text == "" {
		text = concatChunkText(doc.Chunks())
	}
	chunks, err := buildChunks(ctx, uc.chunker, uc.embedder, uc.ids, text)
	if err != nil {
		return ReindexDocumentResult{}, err
	}
	now := uc.clk.Now()
	if err := doc.ReplaceChunks(chunks, now); err != nil {
		return ReindexDocumentResult{}, fmt.Errorf("knowledge: replace chunks: %w", err)
	}
	if uc.retriever != nil {
		if err := uc.retriever.Upsert(ctx, kbID, doc); err != nil {
			return ReindexDocumentResult{}, fmt.Errorf("knowledge: upsert chunks: %w", err)
		}
	}
	if err := doc.MarkIndexed(uc.clk.Now()); err != nil {
		return ReindexDocumentResult{}, fmt.Errorf("knowledge: mark indexed: %w", err)
	}
	if err := uc.docs.Save(ctx, doc); err != nil {
		return ReindexDocumentResult{}, fmt.Errorf("knowledge: save document: %w", err)
	}

	kb, err := uc.kbs.FindByID(ctx, kbID)
	if err != nil {
		return ReindexDocumentResult{}, fmt.Errorf("knowledge: load knowledge base: %w", err)
	}
	if err := kb.AddDocument(doc.ToRef(), uc.clk.Now()); err != nil {
		return ReindexDocumentResult{}, fmt.Errorf("knowledge: refresh document ref: %w", err)
	}
	if err := uc.kbs.Save(ctx, kb); err != nil {
		return ReindexDocumentResult{}, fmt.Errorf("knowledge: save knowledge base: %w", err)
	}

	events := append(doc.PullEvents(), kb.PullEvents()...)
	if err := publishEvents(ctx, uc.bus, events); err != nil {
		return ReindexDocumentResult{}, fmt.Errorf("knowledge: publish events: %w", err)
	}
	return ReindexDocumentResult{DocumentID: doc.ID().String(), ChunkCount: len(doc.Chunks())}, nil
}

// concatChunkText 将文档现有切片文本拼接为重建索引的输入。
func concatChunkText(chunks []*knowledgedomain.Chunk) string {
	var b strings.Builder
	for _, c := range chunks {
		if b.Len() > 0 {
			b.WriteString("\n")
		}
		b.WriteString(c.Content())
	}
	return b.String()
}
