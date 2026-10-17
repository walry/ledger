// Package runtime 存放跨上下文的领域服务（RAG、Agent 执行器、问答助手、数字员工）。
//
// 它们不拥有聚合，只编排底层聚合与端口。
package runtime

import (
	"context"
	"fmt"
	"strings"

	"github.com/walry/ledger/internal/domain/knowledge"
	knowledgesvc "github.com/walry/ledger/internal/domain/knowledge/service"
)

// RAGPipeline 是检索增强生成管线（领域服务）。
type RAGPipeline struct {
	retriever knowledgesvc.Retriever
	reranker  knowledgesvc.Reranker
	topK      int
}

// NewRAGPipeline 创建 RAG 管线。
func NewRAGPipeline(retriever knowledgesvc.Retriever, reranker knowledgesvc.Reranker, topK int) *RAGPipeline {
	if topK <= 0 {
		topK = 5
	}
	return &RAGPipeline{retriever: retriever, reranker: reranker, topK: topK}
}

// Retrieve 检索与查询最相关的切片。
func (p *RAGPipeline) Retrieve(ctx context.Context, query string, kbIDs []knowledge.KnowledgeBaseId, topK int) ([]knowledgesvc.RetrievedChunk, error) {
	if p.retriever == nil || len(kbIDs) == 0 || strings.TrimSpace(query) == "" {
		return nil, nil
	}
	if topK <= 0 {
		topK = p.topK
	}
	chunks, err := p.retriever.Retrieve(ctx, query, kbIDs, topK)
	if err != nil {
		return nil, err
	}
	if p.reranker != nil && len(chunks) > 1 {
		if reranked, rerr := p.reranker.Rerank(ctx, query, chunks, topK); rerr == nil {
			chunks = reranked
		}
	}
	return chunks, nil
}

// BuildContext 将检索结果渲染为提示词上下文。
func (p *RAGPipeline) BuildContext(chunks []knowledgesvc.RetrievedChunk) string {
	if len(chunks) == 0 {
		return ""
	}
	var b strings.Builder
	for i, c := range chunks {
		fmt.Fprintf(&b, "[%d] 来源=%s/%s 相关度=%.3f\n%s\n\n",
			i+1, c.KnowledgeBaseID.String(), c.DocumentID.String(), c.Score, strings.TrimSpace(c.Content))
	}
	return strings.TrimSpace(b.String())
}
