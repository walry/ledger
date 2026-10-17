package qa

import (
	"context"
	"fmt"

	conversationdomain "github.com/walry/ledger/internal/domain/conversation"
	knowledgedomain "github.com/walry/ledger/internal/domain/knowledge"
	runtimedomain "github.com/walry/ledger/internal/domain/runtime"
)

// AskUseCase 是"问答助手"业务用例：编排 runtime.QAAssistant 完成一次问答。
type AskUseCase struct {
	assistant *runtimedomain.QAAssistant
}

// NewAskUseCase 创建问答用例。
func NewAskUseCase(assistant *runtimedomain.QAAssistant) *AskUseCase {
	return &AskUseCase{assistant: assistant}
}

// Execute 执行一次问答（非流式）。
func (uc *AskUseCase) Execute(ctx context.Context, cmd AskCommand) (AskResult, error) {
	in := runtimedomain.RunInput{
		ConversationID:   conversationdomain.NewConversationId(cmd.ConversationID),
		Question:         cmd.Question,
		KnowledgeBaseIDs: toKnowledgeBaseIDs(cmd.KnowledgeBaseIDs),
	}
	res, err := uc.assistant.Ask(ctx, in)
	if err != nil {
		return AskResult{}, fmt.Errorf("qa: ask: %w", err)
	}
	return AskResult{
		Answer: res.Answer,
		TurnID: res.TurnID.String(),
		Tips:   []string{},
	}, nil
}

// toKnowledgeBaseIDs 将字符串 ID 转换为领域标识。
func toKnowledgeBaseIDs(ids []string) []knowledgedomain.KnowledgeBaseId {
	out := make([]knowledgedomain.KnowledgeBaseId, 0, len(ids))
	for _, id := range ids {
		out = append(out, knowledgedomain.NewKnowledgeBaseId(id))
	}
	return out
}
