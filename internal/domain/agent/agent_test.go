package agent

import (
	"testing"
	"time"

	modelconfigdomain "github.com/walry/ledger/internal/domain/modelconfig"
	tooldomain "github.com/walry/ledger/internal/domain/tool"
	"github.com/walry/ledger/pkg/clock"
)

type fixedIDs struct{ n int }

func (f *fixedIDs) New() string { f.n++; return "a-" + string(rune('0'+f.n)) }

func newDraft(t *testing.T, now time.Time) *Agent {
	t.Helper()
	prompt, err := NewSystemPrompt("你是财报分析助手，当前时间 {{current_time}}。")
	if err != nil {
		t.Fatalf("prompt: %v", err)
	}
	ref, err := NewModelConfigRef(modelconfigdomain.NewModelConfigId("mc-1"), "gpt-test", modelconfigdomain.ProviderOpenAI)
	if err != nil {
		t.Fatalf("ref: %v", err)
	}
	ctx, _ := NewAgentContext(nil, 5, 4)
	a, err := NewAgent(NewAgentId("agent-1"), "助手", "desc", prompt, ref, ctx, now)
	if err != nil {
		t.Fatalf("new agent: %v", err)
	}
	return a
}

func TestAgentPublishFreezesConfiguration(t *testing.T) {
	now := time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)
	a := newDraft(t, now)

	skill, _ := NewSkillBinding("skill-1", 1)
	if err := a.AddSkill(skill, now); err != nil {
		t.Fatalf("add skill: %v", err)
	}
	binding, _ := NewToolBinding(tooldomain.NewToolId("tool-1"), 1, map[string]string{"k": "v"})
	if err := a.BindTool(binding, now); err != nil {
		t.Fatalf("bind tool: %v", err)
	}

	if err := a.Publish(now); err != nil {
		t.Fatalf("publish: %v", err)
	}
	if a.Status() != StatusPublished {
		t.Fatalf("expected published, got %s", a.Status())
	}
	events := a.PullEvents()
	if len(events) != 2 || events[0].EventName() != "AgentCreated" || events[1].EventName() != "AgentPublished" {
		t.Fatalf("expected AgentCreated then AgentPublished, got %v", events)
	}

	// 发布后不可编辑
	if err := a.AddSkill(skill, now); err == nil {
		t.Fatalf("expected published agent to reject edits")
	}
	if err := a.BindTool(binding, now); err == nil {
		t.Fatalf("expected published agent to reject tool binding")
	}
	if err := a.Publish(now); err == nil {
		t.Fatalf("expected double publish to fail")
	}
	if got := a.EnabledToolIDs(); len(got) != 1 || got[0] != tooldomain.NewToolId("tool-1") {
		t.Fatalf("unexpected enabled tools: %v", got)
	}
}

func TestFactoryNewVersionCopiesConfiguration(t *testing.T) {
	now := time.Date(2026, 3, 2, 8, 30, 0, 0, time.UTC)
	base := newDraft(t, now)
	skill, _ := NewSkillBinding("skill-1", 2)
	_ = base.AddSkill(skill, now)
	binding, _ := NewToolBinding(tooldomain.NewToolId("tool-1"), 3, nil)
	_ = base.BindTool(binding, now)
	if err := base.Publish(now); err != nil {
		t.Fatalf("publish: %v", err)
	}
	base.PullEvents()

	factory := NewFactory(&fixedIDs{}, clock.Fixed(now))
	next, err := factory.NewVersion(base)
	if err != nil {
		t.Fatalf("new version: %v", err)
	}
	if next.Status() != StatusDraft {
		t.Fatalf("new version should be draft, got %s", next.Status())
	}
	if next.Version() != base.Version()+1 || next.ParentVersion() != base.Version() {
		t.Fatalf("unexpected versioning: v=%d parent=%d", next.Version(), next.ParentVersion())
	}
	if len(next.Skills()) != 1 || len(next.Tools()) != 1 {
		t.Fatalf("expected copied bindings, got skills=%d tools=%d", len(next.Skills()), len(next.Tools()))
	}
	if events := next.PullEvents(); len(events) != 1 {
		t.Fatalf("expected exactly the AgentCreated event, got %d", len(events))
	}
}

func TestSystemPromptRender(t *testing.T) {
	p, err := NewSystemPrompt("你好 {{name}}，现在是 {{time}}。{{name}}")
	if err != nil {
		t.Fatalf("prompt: %v", err)
	}
	if len(p.Variables()) != 2 {
		t.Fatalf("expected 2 unique variables, got %v", p.Variables())
	}
	got := p.Render(map[string]string{"name": "助手", "time": "2026"})
	if got != "你好 助手，现在是 2026。助手" {
		t.Fatalf("unexpected render: %q", got)
	}
}
