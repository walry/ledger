package agent

import "github.com/walry/ledger/internal/domain/knowledge"

// AgentContext 是智能体的配置态运行上下文（值对象）。
type AgentContext struct {
	knowledgeBaseIDs []knowledge.KnowledgeBaseId
	maxSteps         int
	retrievalTopK    int
}

// DefaultAgentContext 返回默认上下文。
func DefaultAgentContext() AgentContext {
	return AgentContext{maxSteps: 6, retrievalTopK: 5}
}

// NewAgentContext 创建上下文并校验不变量。
func NewAgentContext(kbIDs []knowledge.KnowledgeBaseId, maxSteps, topK int) (AgentContext, error) {
	if maxSteps <= 0 {
		maxSteps = 6
	}
	if topK <= 0 {
		topK = 5
	}
	cp := make([]knowledge.KnowledgeBaseId, len(kbIDs))
	copy(cp, kbIDs)
	return AgentContext{knowledgeBaseIDs: cp, maxSteps: maxSteps, retrievalTopK: topK}, nil
}

// KnowledgeBaseIDs 返回绑定的知识库 ID 列表。
func (c AgentContext) KnowledgeBaseIDs() []knowledge.KnowledgeBaseId {
	out := make([]knowledge.KnowledgeBaseId, len(c.knowledgeBaseIDs))
	copy(out, c.knowledgeBaseIDs)
	return out
}

// MaxSteps 返回 ReAct 循环最大步数。
func (c AgentContext) MaxSteps() int {
	if c.maxSteps <= 0 {
		return 6
	}
	return c.maxSteps
}

// RetrievalTopK 返回检索返回条数。
func (c AgentContext) RetrievalTopK() int {
	if c.retrievalTopK <= 0 {
		return 5
	}
	return c.retrievalTopK
}
