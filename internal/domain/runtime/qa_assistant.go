package runtime

import "context"

// QAAssistant 是"问答助手"业务角色对应的领域服务。
//
// 它不拥有聚合，只复用 AgentExecutor 完成 RAG + ReAct 的问答。
type QAAssistant struct {
	executor *AgentExecutor
}

// NewQAAssistant 创建问答助手。
func NewQAAssistant(executor *AgentExecutor) *QAAssistant {
	return &QAAssistant{executor: executor}
}

// Ask 执行一次问答（非流式）。
func (q *QAAssistant) Ask(ctx context.Context, in RunInput) (RunResult, error) {
	return q.executor.Run(ctx, in)
}

// AskStream 执行一次流式问答。
func (q *QAAssistant) AskStream(ctx context.Context, in RunInput) (<-chan StreamEvent, error) {
	return q.executor.Stream(ctx, in)
}
