package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/walry/ledger/internal/domain/conversation"
	"github.com/walry/ledger/internal/domain/modelconfig"
)

// ---------- TruncationStrategy：最简单，直接丢弃旧消息 ----------

// TruncationStrategy 只保留策略要求保留的消息，不做摘要。
type TruncationStrategy struct{}

// Compress 实现 CompressionStrategy。
func (TruncationStrategy) Compress(_ context.Context, _ []conversation.Message, priorSummary string) (CompressedResult, error) {
	return CompressedResult{Summary: priorSummary, Retained: nil}, nil
}

// ---------- SummarizingStrategy：调用 LLM 生成滚动摘要 ----------

const summarizeSystemPrompt = `你是一个对话上下文压缩器。请把给定的对话历史压缩成一段简洁的中文摘要，
用于后续对话的上下文。要求：
1. 逐字保留所有 UUID、文件路径、URL、数字与专有名词，不得改写或省略；
2. 保留用户的关键意图、已确认的结论与进行中的任务；
3. 不要编造历史中不存在的信息；
4. 只输出摘要正文，不要添加任何前后缀。`

// SummarizingStrategy 通过 LLM 生成摘要。
type SummarizingStrategy struct {
	client modelconfig.ClientPort
	model  string
}

// NewSummarizingStrategy 创建摘要策略。
func NewSummarizingStrategy(client modelconfig.ClientPort, model string) *SummarizingStrategy {
	return &SummarizingStrategy{client: client, model: model}
}

// Compress 实现 CompressionStrategy。
func (s *SummarizingStrategy) Compress(ctx context.Context, messages []conversation.Message, priorSummary string) (CompressedResult, error) {
	if s.client == nil {
		return CompressedResult{}, fmt.Errorf("conversation: summarizing strategy has no model client")
	}
	var b strings.Builder
	if priorSummary != "" {
		b.WriteString("已有摘要：\n")
		b.WriteString(priorSummary)
		b.WriteString("\n\n")
	}
	b.WriteString("新增对话：\n")
	for _, m := range messages {
		b.WriteString(string(m.Role()))
		b.WriteString(": ")
		b.WriteString(m.Content())
		if calls := m.ToolCalls(); len(calls) > 0 {
			for _, tc := range calls {
				b.WriteString(fmt.Sprintf("\n  [调用工具 %s: %s]", tc.Name, tc.Arguments))
			}
		}
		b.WriteString("\n")
	}

	resp, err := s.client.Invoke(ctx, modelconfig.Request{
		Model:  s.model,
		System: summarizeSystemPrompt,
		Messages: []modelconfig.Message{
			{Role: modelconfig.RoleUser, Content: b.String()},
		},
		Params: modelconfig.ModelParameters{Temperature: 0, MaxTokens: 1024},
	})
	if err != nil {
		return CompressedResult{}, err
	}
	return CompressedResult{Summary: strings.TrimSpace(resp.Content)}, nil
}

// ---------- SlidingWindowSummaryStrategy：滑动窗口 + 滚动摘要（推荐） ----------

// SlidingWindowSummaryStrategy 在摘要策略之上叠加滚动：把新摘要与旧摘要合并。
type SlidingWindowSummaryStrategy struct {
	inner *SummarizingStrategy
}

// NewSlidingWindowSummaryStrategy 创建滑动窗口摘要策略。
func NewSlidingWindowSummaryStrategy(inner *SummarizingStrategy) *SlidingWindowSummaryStrategy {
	return &SlidingWindowSummaryStrategy{inner: inner}
}

// Compress 实现 CompressionStrategy。
func (s *SlidingWindowSummaryStrategy) Compress(ctx context.Context, messages []conversation.Message, priorSummary string) (CompressedResult, error) {
	if s.inner == nil {
		return CompressedResult{}, fmt.Errorf("conversation: sliding window strategy has no inner strategy")
	}
	return s.inner.Compress(ctx, messages, priorSummary)
}
