package runtime

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	agentdomain "github.com/walry/ledger/internal/domain/agent"
	conversationdomain "github.com/walry/ledger/internal/domain/conversation"
	modelconfigdomain "github.com/walry/ledger/internal/domain/modelconfig"
	tooldomain "github.com/walry/ledger/internal/domain/tool"
	"github.com/walry/ledger/pkg/clock"
)

// ---------- fakes ----------

type seqID struct{ n int }

func (s *seqID) New() string { s.n++; return "id-" + itoa(s.n) }

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}

type memAgentRepo struct {
	m map[agentdomain.AgentId]*agentdomain.Agent
}

func newMemAgentRepo() *memAgentRepo {
	return &memAgentRepo{m: map[agentdomain.AgentId]*agentdomain.Agent{}}
}
func (r *memAgentRepo) Save(_ context.Context, a *agentdomain.Agent) error {
	r.m[a.ID()] = a
	return nil
}
func (r *memAgentRepo) FindByID(_ context.Context, id agentdomain.AgentId) (*agentdomain.Agent, error) {
	a, ok := r.m[id]
	if !ok {
		return nil, agentdomain.ErrNotFound
	}
	return a, nil
}
func (r *memAgentRepo) FindPublished(_ context.Context, id agentdomain.AgentId) (*agentdomain.Agent, error) {
	return r.FindByID(context.Background(), id)
}
func (r *memAgentRepo) List(context.Context, int, int) ([]*agentdomain.Agent, int64, error) {
	return nil, 0, nil
}
func (r *memAgentRepo) Delete(context.Context, agentdomain.AgentId) error { return nil }

type memConvRepo struct {
	m map[conversationdomain.ConversationId]*conversationdomain.Conversation
}

func newMemConvRepo() *memConvRepo {
	return &memConvRepo{m: map[conversationdomain.ConversationId]*conversationdomain.Conversation{}}
}
func (r *memConvRepo) Save(_ context.Context, c *conversationdomain.Conversation) error {
	r.m[c.ID()] = c
	return nil
}
func (r *memConvRepo) FindByID(_ context.Context, id conversationdomain.ConversationId) (*conversationdomain.Conversation, error) {
	c, ok := r.m[id]
	if !ok {
		return nil, conversationdomain.ErrNotFound
	}
	return c, nil
}
func (r *memConvRepo) ListByAgent(context.Context, agentdomain.AgentId, int, int) ([]*conversationdomain.Conversation, int64, error) {
	return nil, 0, nil
}
func (r *memConvRepo) List(context.Context, int, int) ([]*conversationdomain.Conversation, int64, error) {
	return nil, 0, nil
}
func (r *memConvRepo) Delete(context.Context, conversationdomain.ConversationId) error { return nil }

type memToolRepo struct {
	m map[tooldomain.ToolId]*tooldomain.Tool
}

func newMemToolRepo() *memToolRepo {
	return &memToolRepo{m: map[tooldomain.ToolId]*tooldomain.Tool{}}
}
func (r *memToolRepo) Save(_ context.Context, t *tooldomain.Tool) error { r.m[t.ID()] = t; return nil }
func (r *memToolRepo) FindByID(_ context.Context, id tooldomain.ToolId) (*tooldomain.Tool, error) {
	t, ok := r.m[id]
	if !ok {
		return nil, tooldomain.ErrNotFound
	}
	return t, nil
}
func (r *memToolRepo) FindByName(_ context.Context, name string) (*tooldomain.Tool, error) {
	for _, t := range r.m {
		if t.Name() == name {
			return t, nil
		}
	}
	return nil, tooldomain.ErrNotFound
}
func (r *memToolRepo) FindByIDs(_ context.Context, ids []tooldomain.ToolId) ([]*tooldomain.Tool, error) {
	out := make([]*tooldomain.Tool, 0, len(ids))
	for _, id := range ids {
		if t, ok := r.m[id]; ok {
			out = append(out, t)
		}
	}
	return out, nil
}
func (r *memToolRepo) List(context.Context, int, int) ([]*tooldomain.Tool, int64, error) {
	return nil, 0, nil
}
func (r *memToolRepo) Delete(context.Context, tooldomain.ToolId) error { return nil }

type memModelRepo struct {
	m map[modelconfigdomain.ModelConfigId]*modelconfigdomain.ModelConfig
}

func newMemModelRepo() *memModelRepo {
	return &memModelRepo{m: map[modelconfigdomain.ModelConfigId]*modelconfigdomain.ModelConfig{}}
}
func (r *memModelRepo) Save(_ context.Context, c *modelconfigdomain.ModelConfig) error {
	r.m[c.ID()] = c
	return nil
}
func (r *memModelRepo) FindByID(_ context.Context, id modelconfigdomain.ModelConfigId) (*modelconfigdomain.ModelConfig, error) {
	c, ok := r.m[id]
	if !ok {
		return nil, modelconfigdomain.ErrNotFound
	}
	return c, nil
}
func (r *memModelRepo) List(context.Context, int, int) ([]*modelconfigdomain.ModelConfig, int64, error) {
	return nil, 0, nil
}
func (r *memModelRepo) Delete(context.Context, modelconfigdomain.ModelConfigId) error { return nil }

// fakeHandler 记录调用并返回固定结果。
type fakeHandler struct {
	name  string
	calls []json.RawMessage
}

func (h *fakeHandler) Name() string { return h.name }
func (h *fakeHandler) Invoke(_ context.Context, args json.RawMessage) (tooldomain.InvocationResult, error) {
	h.calls = append(h.calls, args)
	return tooldomain.InvocationResult{Content: "42"}, nil
}

type fakeToolFactory struct{ handlers map[string]*fakeHandler }

func (f *fakeToolFactory) Create(name string) (tooldomain.Handler, error) {
	if h, ok := f.handlers[name]; ok {
		return h, nil
	}
	h := &fakeHandler{name: name}
	f.handlers[name] = h
	return h, nil
}
func (f *fakeToolFactory) Has(name string) bool { _, ok := f.handlers[name]; return ok }
func (f *fakeToolFactory) Names() []string {
	out := make([]string, 0, len(f.handlers))
	for n := range f.handlers {
		out = append(out, n)
	}
	return out
}

// scriptedClient 第一轮返回工具调用，第二轮返回最终答复。
type scriptedClient struct {
	invokes int
	streams bool
}

func (c *scriptedClient) Invoke(_ context.Context, _ modelconfigdomain.Request) (modelconfigdomain.Response, error) {
	c.invokes++
	if c.invokes == 1 {
		return modelconfigdomain.Response{
			ToolCalls: []modelconfigdomain.ToolCall{
				{ID: "call-1", Name: "calculator", Arguments: `{"expression":"6*7"}`},
			},
			Usage: modelconfigdomain.Usage{PromptTokens: 10, CompletionTokens: 5, TotalTokens: 15},
		}, nil
	}
	return modelconfigdomain.Response{
		Content:      "营业收入为 42 万元。",
		FinishReason: "stop",
		Usage:        modelconfigdomain.Usage{PromptTokens: 20, CompletionTokens: 8, TotalTokens: 28},
	}, nil
}

func (c *scriptedClient) Stream(ctx context.Context, _ modelconfigdomain.Request) (<-chan modelconfigdomain.Chunk, error) {
	ch := make(chan modelconfigdomain.Chunk, 8)
	c.invokes++
	first := c.invokes == 1
	go func() {
		defer close(ch)
		if first {
			ch <- modelconfigdomain.Chunk{Delta: "我来计算："}
			ch <- modelconfigdomain.Chunk{ToolCallDelta: &modelconfigdomain.ToolCall{ID: "call-1", Name: "calculator", Arguments: `{"expression":"6*7"}`}}
			ch <- modelconfigdomain.Chunk{Done: true}
			return
		}
		for _, d := range []string{"营业收入", "为 42", " 万元。"} {
			ch <- modelconfigdomain.Chunk{Delta: d}
		}
		ch <- modelconfigdomain.Chunk{Done: true, Usage: &modelconfigdomain.Usage{TotalTokens: 30}}
	}()
	return ch, nil
}

type fakeClientFactory struct{ client *scriptedClient }

func (f *fakeClientFactory) Create(*modelconfigdomain.ModelConfig) (modelconfigdomain.ClientPort, error) {
	return f.client, nil
}
func (f *fakeClientFactory) OnConfigUpdated(context.Context, *modelconfigdomain.ModelConfig) error {
	return nil
}

// ---------- test fixture ----------

type fixture struct {
	executor *AgentExecutor
	agents   *memAgentRepo
	convs    *memConvRepo
	tools    *memToolRepo
	models   *memModelRepo
	handlers *fakeToolFactory
	client   *scriptedClient
	convID   conversationdomain.ConversationId
	agentID  agentdomain.AgentId
	now      time.Time
}

func newFixture(t *testing.T, streaming bool) *fixture {
	t.Helper()
	ctx := context.Background()
	ids := &seqID{}
	clk := clock.Fixed(time.Date(2026, 1, 2, 15, 4, 5, 0, time.UTC))
	now := clk.Now()

	agents := newMemAgentRepo()
	convs := newMemConvRepo()
	tools := newMemToolRepo()
	models := newMemModelRepo()
	handlers := &fakeToolFactory{handlers: map[string]*fakeHandler{}}
	client := &scriptedClient{streams: streaming}

	// 模型配置
	mc, err := modelconfigdomain.NewModelConfig(
		modelconfigdomain.NewModelConfigId("mc-1"), "test", modelconfigdomain.ProviderOpenAI,
		"http://localhost", "key", "gpt-test", modelconfigdomain.DefaultParameters(), now)
	if err != nil {
		t.Fatalf("model config: %v", err)
	}
	if err := models.Save(ctx, mc); err != nil {
		t.Fatalf("save model config: %v", err)
	}

	// 工具
	toolSchema := tooldomain.ObjectSchema(map[string]tooldomain.JSONSchema{
		"expression": {Type: "string"},
	}, "expression")
	src, _ := tooldomain.NewToolSource(tooldomain.SourceBuiltin, "", nil)
	href, _ := tooldomain.NewHandlerRef("calculator")
	toolAgg, err := tooldomain.NewTool(tooldomain.NewToolId("tool-1"), "calculator", "计算", toolSchema, src, href, 5*time.Second, now)
	if err != nil {
		t.Fatalf("tool: %v", err)
	}
	if err := tools.Save(ctx, toolAgg); err != nil {
		t.Fatalf("save tool: %v", err)
	}

	// 智能体：绑定工具并发布
	prompt, _ := agentdomain.NewSystemPrompt("你是财报分析助手。")
	ref, _ := agentdomain.NewModelConfigRef(mc.ID(), mc.Model(), mc.Provider())
	actx, _ := agentdomain.NewAgentContext(nil, 4, 3)
	agentAgg, err := agentdomain.NewAgent(agentdomain.NewAgentId("agent-1"), "财报助手", "", prompt, ref, actx, now)
	if err != nil {
		t.Fatalf("agent: %v", err)
	}
	binding, _ := agentdomain.NewToolBinding(toolAgg.ID(), 1, nil)
	if err := agentAgg.BindTool(binding, now); err != nil {
		t.Fatalf("bind tool: %v", err)
	}
	if err := agentAgg.Publish(now); err != nil {
		t.Fatalf("publish: %v", err)
	}
	agentAgg.PullEvents()
	if err := agents.Save(ctx, agentAgg); err != nil {
		t.Fatalf("save agent: %v", err)
	}

	// 会话
	conv, err := conversationdomain.NewConversation(conversationdomain.NewConversationId("conv-1"), "测试会话", agentAgg.ID(), agentAgg.Version(), now)
	if err != nil {
		t.Fatalf("conversation: %v", err)
	}
	conv.PullEvents()
	if err := convs.Save(ctx, conv); err != nil {
		t.Fatalf("save conversation: %v", err)
	}

	exec := NewAgentExecutor(AgentExecutorConfig{
		Agents:        agents,
		Conversations: convs,
		Tools:         tools,
		ToolFactory:   handlers,
		Models:        &fakeClientFactory{client: client},
		ModelConfigs:  models,
		RAG:           nil,
		Compressor:    nil,
		Events:        nil,
		IDs:           ids,
		Clock:         clk,
	})

	return &fixture{
		executor: exec, agents: agents, convs: convs, tools: tools, models: models,
		handlers: handlers, client: client, convID: conv.ID(), agentID: agentAgg.ID(), now: now,
	}
}

// ---------- tests ----------

func TestAgentExecutorRun_ExecutesToolThenAnswers(t *testing.T) {
	fx := newFixture(t, false)
	ctx := context.Background()

	res, err := fx.executor.Run(ctx, RunInput{ConversationID: fx.convID, Question: "营业收入是多少？"})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if res.Answer != "营业收入为 42 万元。" {
		t.Fatalf("unexpected answer: %q", res.Answer)
	}
	if res.Steps != 2 {
		t.Fatalf("expected 2 steps, got %d", res.Steps)
	}
	if res.Usage.TotalTokens != 43 {
		t.Fatalf("expected aggregated usage 43, got %d", res.Usage.TotalTokens)
	}
	if len(fx.handlers.handlers["calculator"].calls) != 1 {
		t.Fatalf("expected calculator invoked once")
	}
	if got := string(fx.handlers.handlers["calculator"].calls[0]); got != `{"expression":"6*7"}` {
		t.Fatalf("unexpected tool args: %s", got)
	}

	// 会话应被持久化，且包含完成的轮次与工具结果消息
	conv, err := fx.convs.FindByID(ctx, fx.convID)
	if err != nil {
		t.Fatalf("reload conversation: %v", err)
	}
	if len(conv.Turns()) != 1 {
		t.Fatalf("expected 1 turn, got %d", len(conv.Turns()))
	}
	turn := conv.Turns()[0]
	if turn.Status() != conversationdomain.TurnCompleted {
		t.Fatalf("expected completed turn, got %s", turn.Status())
	}
	roles := make([]conversationdomain.Role, 0, len(turn.Messages()))
	for _, m := range turn.Messages() {
		roles = append(roles, m.Role())
	}
	want := []conversationdomain.Role{
		conversationdomain.RoleUser,
		conversationdomain.RoleAssistant,
		conversationdomain.RoleTool,
		conversationdomain.RoleAssistant,
	}
	if len(roles) != len(want) {
		t.Fatalf("unexpected message roles: %v", roles)
	}
	for i := range want {
		if roles[i] != want[i] {
			t.Fatalf("role[%d] = %s, want %s", i, roles[i], want[i])
		}
	}
}

func TestAgentExecutorStream_EmitsDeltasAndToolEvents(t *testing.T) {
	fx := newFixture(t, true)
	ctx := context.Background()

	events, err := fx.executor.Stream(ctx, RunInput{ConversationID: fx.convID, Question: "营业收入是多少？"})
	if err != nil {
		t.Fatalf("stream: %v", err)
	}
	var (
		deltas     string
		toolCalls  int
		toolResult int
		done       *StreamEvent
		sawErr     error
	)
	for ev := range events {
		switch ev.Type {
		case EventDelta:
			deltas += ev.Delta
		case EventToolCall:
			toolCalls++
		case EventToolResult:
			toolResult++
		case EventError:
			sawErr = ev.Err
		case EventDone:
			e := ev
			done = &e
		}
	}
	if sawErr != nil {
		t.Fatalf("stream error: %v", sawErr)
	}
	if done == nil {
		t.Fatalf("expected a done event")
	}
	if done.Answer != "营业收入为 42 万元。" {
		t.Fatalf("unexpected final answer: %q", done.Answer)
	}
	if toolCalls != 1 || toolResult != 1 {
		t.Fatalf("expected 1 tool call and 1 result, got %d/%d", toolCalls, toolResult)
	}
	if len(deltas) == 0 {
		t.Fatalf("expected streamed deltas")
	}
	if done.Usage.TotalTokens != 30 {
		t.Fatalf("expected usage 30, got %d", done.Usage.TotalTokens)
	}
}

func TestAgentExecutorRun_RejectsArchivedConversation(t *testing.T) {
	fx := newFixture(t, false)
	ctx := context.Background()
	conv, _ := fx.convs.FindByID(ctx, fx.convID)
	conv.Archive(fx.now)
	if _, err := fx.executor.Run(ctx, RunInput{ConversationID: fx.convID, Question: "再来一次"}); err == nil {
		t.Fatalf("expected error for archived conversation")
	}
}
