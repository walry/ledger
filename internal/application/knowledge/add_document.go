package knowledge

import (
	"context"
	"errors"
	"fmt"

	"github.com/walry/ledger/internal/domain/ddd"
	knowledgedomain "github.com/walry/ledger/internal/domain/knowledge"
	knowledgesvc "github.com/walry/ledger/internal/domain/knowledge/service"
	"github.com/walry/ledger/pkg/clock"
	"github.com/walry/ledger/pkg/idgen"
)

// AddDocumentUseCase 向知识库新增文档：切片、向量化、写入向量库并建立索引。
type AddDocumentUseCase struct {
	kbs       knowledgedomain.KnowledgeBaseRepository
	docs      knowledgedomain.DocumentRepository
	chunker   knowledgesvc.Chunker
	embedder  knowledgesvc.Embedder
	retriever knowledgesvc.Retriever
	ids       idgen.Generator
	clk       clock.Clock
	bus       ddd.EventBus
}

// NewAddDocumentUseCase 创建新增文档用例。
func NewAddDocumentUseCase(
	kbs knowledgedomain.KnowledgeBaseRepository,
	docs knowledgedomain.DocumentRepository,
	chunker knowledgesvc.Chunker,
	embedder knowledgesvc.Embedder,
	retriever knowledgesvc.Retriever,
	ids idgen.Generator,
	clk clock.Clock,
	bus ddd.EventBus,
) *AddDocumentUseCase {
	return &AddDocumentUseCase{
		kbs: kbs, docs: docs, chunker: chunker, embedder: embedder,
		retriever: retriever, ids: ids, clk: clk, bus: bus,
	}
}

// Execute 执行新增。
func (uc *AddDocumentUseCase) Execute(ctx context.Context, cmd AddDocumentCommand) (AddDocumentResult, error) {
	kbID := knowledgedomain.NewKnowledgeBaseId(cmd.KnowledgeBaseID)
	kb, err := uc.kbs.FindByID(ctx, kbID)
	if err != nil {
		if errors.Is(err, knowledgedomain.ErrNotFound) {
			return AddDocumentResult{}, fmt.Errorf("knowledge: knowledge base %s not found: %w", cmd.KnowledgeBaseID, err)
		}
		return AddDocumentResult{}, fmt.Errorf("knowledge: load knowledge base: %w", err)
	}

	now := uc.clk.Now()
	source := knowledgedomain.SourceRef{
		Kind:  cmd.SourceKind,
		RefID: cmd.SourceRefID,
		Title: cmd.Title,
	}
	doc, err := knowledgedomain.NewDocument(
		knowledgedomain.DocumentId(uc.ids.New()),
		kbID,
		source,
		cmd.Title,
		now,
	)
	if err != nil {
		return AddDocumentResult{}, fmt.Errorf("knowledge: create document: %w", err)
	}

	chunks, err := buildChunks(ctx, uc.chunker, uc.embedder, uc.ids, cmd.Text)
	if err != nil {
		return AddDocumentResult{}, err
	}
	if err := doc.ReplaceChunks(chunks, now); err != nil {
		return AddDocumentResult{}, fmt.Errorf("knowledge: replace chunks: %w", err)
	}
	if uc.retriever != nil {
		if err := uc.retriever.Upsert(ctx, kbID, doc); err != nil {
			return AddDocumentResult{}, fmt.Errorf("knowledge: upsert chunks: %w", err)
		}
	}
	if err := doc.MarkIndexed(uc.clk.Now()); err != nil {
		return AddDocumentResult{}, fmt.Errorf("knowledge: mark indexed: %w", err)
	}
	if err := uc.docs.Save(ctx, doc); err != nil {
		return AddDocumentResult{}, fmt.Errorf("knowledge: save document: %w", err)
	}
	if err := kb.AddDocument(doc.ToRef(), uc.clk.Now()); err != nil {
		return AddDocumentResult{}, fmt.Errorf("knowledge: register document ref: %w", err)
	}
	if err := uc.kbs.Save(ctx, kb); err != nil {
		return AddDocumentResult{}, fmt.Errorf("knowledge: save knowledge base: %w", err)
	}

	events := append(doc.PullEvents(), kb.PullEvents()...)
	if err := publishEvents(ctx, uc.bus, events); err != nil {
		return AddDocumentResult{}, fmt.Errorf("knowledge: publish events: %w", err)
	}
	return AddDocumentResult{DocumentID: doc.ID().String(), ChunkCount: len(doc.Chunks())}, nil
}

// buildChunks 切片文本并按需向量化，产出文档切片实体。被 add/reindex 复用。
func buildChunks(ctx context.Context, chunker knowledgesvc.Chunker, embedder knowledgesvc.Embedder, ids idgen.Generator, text string) ([]*knowledgedomain.Chunk, error) {
	raw, err := chunker.Chunk(ctx, text, knowledgesvc.ChunkOptions{})
	if err != nil {
		return nil, fmt.Errorf("knowledge: chunk text: %w", err)
	}
	if len(raw) == 0 {
		return nil, errors.New("knowledge: text produced no chunks")
	}
	contents := make([]string, len(raw))
	for i, rc := range raw {
		contents[i] = rc.Content
	}
	var embeddings []knowledgedomain.Embedding
	if embedder != nil {
		embeddings, err = embedder.Embed(ctx, contents)
		if err != nil {
			return nil, fmt.Errorf("knowledge: embed chunks: %w", err)
		}
		if len(embeddings) != len(raw) {
			return nil, fmt.Errorf("knowledge: embedder returned %d vectors for %d chunks", len(embeddings), len(raw))
		}
	}
	chunks := make([]*knowledgedomain.Chunk, 0, len(raw))
	for i, rc := range raw {
		chunk, err := knowledgedomain.NewChunk(knowledgedomain.ChunkId(ids.New()), rc.Index, rc.Content, rc.Metadata)
		if err != nil {
			return nil, fmt.Errorf("knowledge: build chunk: %w", err)
		}
		if rc.Tokens > 0 {
			chunk.SetTokenCount(rc.Tokens)
		}
		if embeddings != nil {
			chunk.SetEmbedding(embeddings[i])
		}
		chunks = append(chunks, chunk)
	}
	return chunks, nil
}
