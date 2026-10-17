package qa

import (
	"context"
	"fmt"

	conversationdomain "github.com/walry/ledger/internal/domain/conversation"
	runtimedomain "github.com/walry/ledger/internal/domain/runtime"
)

// AskStreamUseCase 是"问答助手"流式业务用例。
type AskStreamUseCase struct {
	assistant *runtimedomain.QAAssistant
}

// NewAskStreamUseCase 创建流式问答用例。
func NewAskStreamUseCase(assistant *runtimedomain.QAAssistant) *AskStreamUseCase {
	return &AskStreamUseCase{assistant: assistant}
}

// ExecuteStream 执行流式问答，返回运行时事件流。
func (uc *AskStreamUseCase) ExecuteStream(ctx context.Context, cmd AskCommand) (<-chan runtimedomain.StreamEvent, error) {
	in := runtimedomain.RunInput{
		ConversationID:   conversationdomain.NewConversationId(cmd.ConversationID),
		Question:         cmd.Question,
		KnowledgeBaseIDs: toKnowledgeBaseIDs(cmd.KnowledgeBaseIDs),
	}
	ch, err := uc.assistant.AskStream(ctx, in)
	if err != nil {
		return nil, fmt.Errorf("qa: ask stream: %w", err)
	}
	return ch, nil
}
