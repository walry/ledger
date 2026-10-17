package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/walry/ledger/internal/domain/agent"
	"github.com/walry/ledger/internal/domain/conversation"
	convsrv "github.com/walry/ledger/internal/domain/conversation/service"
	"github.com/walry/ledger/internal/domain/ddd"
	"github.com/walry/ledger/internal/domain/knowledge"
	knowledgesvc "github.com/walry/ledger/internal/domain/knowledge/service"
	"github.com/walry/ledger/internal/domain/modelconfig"
	"github.com/walry/ledger/internal/domain/tool"
	"github.com/walry/ledger/pkg/clock"
	"github.com/walry/ledger/pkg/idgen"
)

// AgentExecutorConfig 是执行器依赖。
type AgentExecutorConfig struct {
	Agents        agent.Repository
	Conversations conversation.Repository
	Tools         tool.Repository
	ToolFactory   tool.HandlerFactory
	Models        modelconfig.ClientFactory
	ModelConfigs  modelconfig.Repository
	RAG           *RAGPipeline
	Compressor    *convsrv.ContextCompressor
	Events        ddd.EventBus
	IDs           idgen.Generator
	Clock         clock.Clock
}

// AgentExecutor 是 ReAct 循环执行器（领域服务）。
type AgentExecutor struct {
	cfg AgentExecutorConfig
}

// NewAgentExecutor 创建执行器。
func NewAgentExecutor(cfg AgentExecutorConfig) *AgentExecutor {
	if cfg.Clock == nil {
		cfg.Clock = clock.System()
	}
	return &AgentExecutor{cfg: cfg}
}

// RunInput 是一次执行的输入。
type RunInput struct {
	ConversationID   conversation.ConversationId
	Question         string
	KnowledgeBaseIDs []knowledge.KnowledgeBaseId // 为空时使用 Agent 绑定的知识库
}

// RunResult 是一次执行的输出。
type RunResult struct {
	TurnID    conversation.TurnId
	Answer    string
	Steps     int
	Usage     modelconfig.Usage
	Retrieved []knowledgesvc.RetrievedChunk
}

// StreamEventType 是流式事件类型。
type StreamEventType string

const (
	// EventDelta 文本增量。
	EventDelta StreamEventType = "delta"
	// EventToolCall 工具调用。
	EventToolCall StreamEventType = "tool_call"
	// EventToolResult 工具结果。
	EventToolResult StreamEventType = "tool_result"
	// EventDone 完成。
	EventDone StreamEventType = "done"
	// EventError 错误。
	EventError StreamEventType = "error"
)

// StreamEvent 是流式事件。
type StreamEvent struct {
	Type       StreamEventType
	Delta      string
	ToolCallID string
	ToolName   string
	ToolResult string
	Answer     string
	Usage      modelconfig.Usage
	Err        error
}

// Run 执行一次完整的 ReAct 循环（非流式）。
func (e *AgentExecutor) Run(ctx context.Context, in RunInput) (RunResult, error) {
	conv, a, mc, client, err := e.prepare(ctx, in)
	if err != nil {
		return RunResult{}, err
	}
	now := e.cfg.Clock.Now()
	turn, err := conv.StartTurn(e.cfg.IDs, now)
	if err != nil {
		return RunResult{}, err
	}
	turn.AppendMessage(conversation.NewUserMessage(in.Question, now))

	e.maybeCompress(ctx, conv)
	chunks, contextText := e.retrieve(ctx, a, in.Question, in.KnowledgeBaseIDs)
	schemas, toolIndex := e.buildToolSchemas(ctx, a)
	req := modelconfig.Request{
		Model:    a.ModelRef().Model,
		System:   e.systemPrompt(a, contextText, now),
		Messages: e.buildHistory(conv),
		Tools:    schemas,
		Params:   mc.Parameters(),
	}

	maxSteps := a.Context().MaxSteps()
	var usage modelconfig.Usage
	answer := ""
	modelCalls := 0
	for step := 0; step < maxSteps; step++ {
		resp, err := client.Invoke(ctx, req)
		modelCalls++
		if err != nil {
			e.failTurn(ctx, conv, err)
			return RunResult{}, err
		}
		usage = addUsage(usage, resp.Usage)
		if len(resp.ToolCalls) == 0 {
			answer = strings.TrimSpace(resp.Content)
			turn.AppendMessage(conversation.NewAssistantMessage(resp.Content, nil, e.cfg.Clock.Now()))
			break
		}
		turn.AppendMessage(conversation.NewAssistantMessage(resp.Content, toConversationCalls(resp.ToolCalls), e.cfg.Clock.Now()))
		req.Messages = append(req.Messages, modelconfig.Message{
			Role: modelconfig.RoleAssistant, Content: resp.Content, ToolCalls: resp.ToolCalls,
		})
		for _, tc := range resp.ToolCalls {
			content, _ := e.invokeTool(ctx, toolIndex, tc)
			turn.AppendMessage(conversation.NewToolMessage(tc.ID, tc.Name, content, e.cfg.Clock.Now()))
			req.Messages = append(req.Messages, modelconfig.Message{
				Role: modelconfig.RoleTool, ToolCallID: tc.ID, Name: tc.Name, Content: content,
			})
		}
	}
	if answer == "" {
		answer = strings.TrimSpace(turn.LastAssistantContent())
	}
	conv.CompleteTurn(e.cfg.Clock.Now())
	if err := e.save(ctx, conv); err != nil {
		return RunResult{}, err
	}
	return RunResult{TurnID: turn.ID(), Answer: answer, Steps: modelCalls, Usage: usage, Retrieved: chunks}, nil
}

// Stream 执行 ReAct 循环并以流式事件返回结果。
func (e *AgentExecutor) Stream(ctx context.Context, in RunInput) (<-chan StreamEvent, error) {
	conv, a, mc, client, err := e.prepare(ctx, in)
	if err != nil {
		return nil, err
	}
	out := make(chan StreamEvent, 32)
	go func() {
		defer close(out)
		now := e.cfg.Clock.Now()
		turn, err := conv.StartTurn(e.cfg.IDs, now)
		if err != nil {
			out <- StreamEvent{Type: EventError, Err: err}
			return
		}
		turn.AppendMessage(conversation.NewUserMessage(in.Question, now))

		e.maybeCompress(ctx, conv)
		_, contextText := e.retrieve(ctx, a, in.Question, in.KnowledgeBaseIDs)
		schemas, toolIndex := e.buildToolSchemas(ctx, a)
		req := modelconfig.Request{
			Model:    a.ModelRef().Model,
			System:   e.systemPrompt(a, contextText, now),
			Messages: e.buildHistory(conv),
			Tools:    schemas,
			Params:   mc.Parameters(),
			Stream:   true,
		}

		maxSteps := a.Context().MaxSteps()
		var usage modelconfig.Usage
		finalAnswer := ""
		for step := 0; step < maxSteps; step++ {
			stream, serr := client.Stream(ctx, req)
			if serr != nil {
				e.failTurn(ctx, conv, serr)
				out <- StreamEvent{Type: EventError, Err: serr}
				return
			}
			var content strings.Builder
			var calls []modelconfig.ToolCall
			for chunk := range stream {
				if chunk.Err != nil {
					e.failTurn(ctx, conv, chunk.Err)
					out <- StreamEvent{Type: EventError, Err: chunk.Err}
					return
				}
				if chunk.Delta != "" {
					content.WriteString(chunk.Delta)
					out <- StreamEvent{Type: EventDelta, Delta: chunk.Delta}
				}
				if chunk.ToolCallDelta != nil {
					calls = mergeToolCall(calls, *chunk.ToolCallDelta)
				}
				if chunk.Usage != nil {
					usage = addUsage(usage, *chunk.Usage)
				}
			}
			if len(calls) == 0 {
				finalAnswer = strings.TrimSpace(content.String())
				turn.AppendMessage(conversation.NewAssistantMessage(content.String(), nil, e.cfg.Clock.Now()))
				break
			}
			turn.AppendMessage(conversation.NewAssistantMessage(content.String(), toConversationCalls(calls), e.cfg.Clock.Now()))
			req.Messages = append(req.Messages, modelconfig.Message{
				Role: modelconfig.RoleAssistant, Content: content.String(), ToolCalls: calls,
			})
			for _, tc := range calls {
				out <- StreamEvent{Type: EventToolCall, ToolCallID: tc.ID, ToolName: tc.Name}
				resultText, _ := e.invokeTool(ctx, toolIndex, tc)
				turn.AppendMessage(conversation.NewToolMessage(tc.ID, tc.Name, resultText, e.cfg.Clock.Now()))
				out <- StreamEvent{Type: EventToolResult, ToolCallID: tc.ID, ToolName: tc.Name, ToolResult: resultText}
				req.Messages = append(req.Messages, modelconfig.Message{
					Role: modelconfig.RoleTool, ToolCallID: tc.ID, Name: tc.Name, Content: resultText,
				})
			}
		}
		if finalAnswer == "" {
			finalAnswer = strings.TrimSpace(turn.LastAssistantContent())
		}
		conv.CompleteTurn(e.cfg.Clock.Now())
		if err := e.save(ctx, conv); err != nil {
			out <- StreamEvent{Type: EventError, Err: err}
			return
		}
		out <- StreamEvent{Type: EventDone, Answer: finalAnswer, Usage: usage}
	}()
	return out, nil
}

// ---------- 内部辅助 ----------

func (e *AgentExecutor) prepare(ctx context.Context, in RunInput) (*conversation.Conversation, *agent.Agent, *modelconfig.ModelConfig, modelconfig.ClientPort, error) {
	if in.ConversationID.IsZero() {
		return nil, nil, nil, nil, errors.New("runtime: conversation id is required")
	}
	conv, err := e.cfg.Conversations.FindByID(ctx, in.ConversationID)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	if err := conv.EnsureActive(); err != nil {
		return nil, nil, nil, nil, err
	}
	a, err := e.cfg.Agents.FindByID(ctx, conv.AgentID())
	if err != nil {
		return nil, nil, nil, nil, err
	}
	mc, err := e.cfg.ModelConfigs.FindByID(ctx, a.ModelRef().ConfigID)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	if !mc.Enabled() {
		return nil, nil, nil, nil, fmt.Errorf("runtime: model config %s is disabled", mc.ID())
	}
	client, err := e.cfg.Models.Create(mc)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	return conv, a, mc, client, nil
}

func (e *AgentExecutor) systemPrompt(a *agent.Agent, contextText string, now time.Time) string {
	vars := map[string]string{
		"knowledge_context": contextText,
		"current_time":      now.Format(time.RFC3339),
		"agent_name":        a.Name(),
	}
	system := a.SystemPrompt().Render(vars)
	if contextText != "" && !strings.Contains(a.SystemPrompt().Template(), "knowledge_context") {
		system += "\n\n以下是从知识库检索到的参考资料，请优先依据资料回答，并在引用时注明来源编号：\n" + contextText
	}
	return system
}

func (e *AgentExecutor) buildHistory(conv *conversation.Conversation) []modelconfig.Message {
	msgs := make([]modelconfig.Message, 0)
	for _, t := range conv.Turns() {
		for _, m := range t.Messages() {
			switch m.Role() {
			case conversation.RoleSystem:
				continue
			case conversation.RoleUser:
				msgs = append(msgs, modelconfig.Message{Role: modelconfig.RoleUser, Content: m.Content()})
			case conversation.RoleAssistant:
				msgs = append(msgs, modelconfig.Message{Role: modelconfig.RoleAssistant, Content: m.Content(), ToolCalls: toModelCalls(m.ToolCalls())})
			case conversation.RoleTool:
				msgs = append(msgs, modelconfig.Message{Role: modelconfig.RoleTool, Content: m.Content(), Name: m.Name(), ToolCallID: m.ToolCallID()})
			}
		}
	}
	return msgs
}

func (e *AgentExecutor) retrieve(ctx context.Context, a *agent.Agent, question string, override []knowledge.KnowledgeBaseId) ([]knowledgesvc.RetrievedChunk, string) {
	if e.cfg.RAG == nil {
		return nil, ""
	}
	kbIDs := override
	if len(kbIDs) == 0 {
		kbIDs = a.KnowledgeBaseIDs()
	}
	chunks, err := e.cfg.RAG.Retrieve(ctx, question, kbIDs, a.Context().RetrievalTopK())
	if err != nil {
		return nil, ""
	}
	return chunks, e.cfg.RAG.BuildContext(chunks)
}

func (e *AgentExecutor) buildToolSchemas(ctx context.Context, a *agent.Agent) ([]modelconfig.ToolSchema, map[string]tool.ToolId) {
	ids := a.EnabledToolIDs()
	index := map[string]tool.ToolId{}
	if len(ids) == 0 || e.cfg.Tools == nil {
		return nil, index
	}
	tools, err := e.cfg.Tools.FindByIDs(ctx, ids)
	if err != nil {
		return nil, index
	}
	schemas := make([]modelconfig.ToolSchema, 0, len(tools))
	for _, t := range tools {
		if !t.Enabled() {
			continue
		}
		index[t.Name()] = t.ID()
		schemas = append(schemas, modelconfig.ToolSchema{
			Name:        t.Name(),
			Description: t.Description(),
			Parameters:  t.Schema().Raw(),
		})
	}
	return schemas, index
}

func (e *AgentExecutor) invokeTool(ctx context.Context, index map[string]tool.ToolId, tc modelconfig.ToolCall) (string, bool) {
	id, ok := index[tc.Name]
	if !ok {
		return fmt.Sprintf("工具 %s 未绑定或未启用，无法调用。", tc.Name), true
	}
	t, err := e.cfg.Tools.FindByID(ctx, id)
	if err != nil {
		return fmt.Sprintf("加载工具 %s 失败: %v", tc.Name, err), true
	}
	if e.cfg.ToolFactory == nil {
		return fmt.Sprintf("工具执行器未配置"), true
	}
	h, err := e.cfg.ToolFactory.Create(t.Handler().Name())
	if err != nil {
		return fmt.Sprintf("创建工具处理器失败: %v", err), true
	}
	tctx := ctx
	if t.Timeout() > 0 {
		var cancel context.CancelFunc
		tctx, cancel = context.WithTimeout(ctx, t.Timeout())
		defer cancel()
	}
	args := json.RawMessage(tc.Arguments)
	if len(args) == 0 {
		args = json.RawMessage(`{}`)
	}
	res, err := h.Invoke(tctx, args)
	if err != nil {
		return fmt.Sprintf("工具 %s 执行失败: %v", tc.Name, err), true
	}
	return res.Content, res.IsError
}

func (e *AgentExecutor) maybeCompress(ctx context.Context, conv *conversation.Conversation) {
	if e.cfg.Compressor == nil {
		return
	}
	msgs := e.buildConversationMessages(conv)
	if !e.cfg.Compressor.ShouldCompress(len(msgs)) {
		return
	}
	prior := ""
	if s := conv.Summary(); s != nil {
		prior = s.Content()
	}
	res, err := e.cfg.Compressor.Compress(ctx, msgs, prior)
	if err != nil || res.Summary == "" {
		return
	}
	sum, err := conversation.NewConversationSummary(res.Summary, len(conv.Turns())-1, e.cfg.Clock.Now())
	if err != nil {
		return
	}
	conv.SetSummary(sum, e.cfg.Clock.Now())
}

func (e *AgentExecutor) buildConversationMessages(conv *conversation.Conversation) []conversation.Message {
	msgs := make([]conversation.Message, 0)
	for _, t := range conv.Turns() {
		msgs = append(msgs, t.Messages()...)
	}
	return msgs
}

func (e *AgentExecutor) failTurn(ctx context.Context, conv *conversation.Conversation, cause error) {
	conv.FailTurn(cause.Error(), e.cfg.Clock.Now())
	_ = e.save(ctx, conv)
}

func (e *AgentExecutor) save(ctx context.Context, conv *conversation.Conversation) error {
	if err := e.cfg.Conversations.Save(ctx, conv); err != nil {
		return err
	}
	if e.cfg.Events != nil {
		if events := conv.PullEvents(); len(events) > 0 {
			return e.cfg.Events.Publish(ctx, events...)
		}
	}
	return nil
}

func toConversationCalls(calls []modelconfig.ToolCall) []conversation.ToolCall {
	out := make([]conversation.ToolCall, 0, len(calls))
	for _, c := range calls {
		out = append(out, conversation.ToolCall{ID: c.ID, Name: c.Name, Arguments: c.Arguments})
	}
	return out
}

func toModelCalls(calls []conversation.ToolCall) []modelconfig.ToolCall {
	out := make([]modelconfig.ToolCall, 0, len(calls))
	for _, c := range calls {
		out = append(out, modelconfig.ToolCall{ID: c.ID, Name: c.Name, Arguments: c.Arguments})
	}
	return out
}

func mergeToolCall(calls []modelconfig.ToolCall, delta modelconfig.ToolCall) []modelconfig.ToolCall {
	for i, c := range calls {
		if c.ID != "" && c.ID == delta.ID {
			if delta.Name != "" {
				calls[i].Name = delta.Name
			}
			if delta.Arguments != "" {
				calls[i].Arguments = delta.Arguments
			}
			return calls
		}
	}
	return append(calls, delta)
}

func addUsage(a, b modelconfig.Usage) modelconfig.Usage {
	return modelconfig.Usage{
		PromptTokens:     a.PromptTokens + b.PromptTokens,
		CompletionTokens: a.CompletionTokens + b.CompletionTokens,
		TotalTokens:      a.TotalTokens + b.TotalTokens,
	}
}
