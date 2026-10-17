package agent

import (
	"github.com/walry/ledger/pkg/clock"
	"github.com/walry/ledger/pkg/idgen"
)

// Factory 负责创建 Agent 聚合。
type Factory struct {
	ids   idgen.Generator
	clock clock.Clock
}

// NewFactory 创建工厂。
func NewFactory(ids idgen.Generator, clk clock.Clock) *Factory {
	return &Factory{ids: ids, clock: clk}
}

// Create 创建草稿智能体。
func (f *Factory) Create(name, description string, prompt SystemPrompt, modelRef ModelConfigRef, ctx AgentContext) (*Agent, error) {
	return NewAgent(AgentId(f.ids.New()), name, description, prompt, modelRef, ctx, f.clock.Now())
}

// NewVersion 基于已有智能体派生一个新版本（草稿，版本号 +1）。
func (f *Factory) NewVersion(base *Agent) (*Agent, error) {
	// 复用聚合的只读访问器，仅复制配置（不复制事件）。
	next, err := NewAgent(
		AgentId(f.ids.New()),
		base.Name(),
		base.Description(),
		base.SystemPrompt(),
		base.ModelRef(),
		base.Context(),
		f.clock.Now(),
	)
	if err != nil {
		return nil, err
	}
	for _, s := range base.Skills() {
		_ = next.AddSkill(s, f.clock.Now())
	}
	for _, t := range base.Tools() {
		_ = next.BindTool(t, f.clock.Now())
	}
	next.version = base.Version() + 1
	next.parentVersion = base.Version()
	// 重新记录事件，使其版本号反映派生后的版本（NewAgent 默认记录的是 version=1）。
	next.events = nil
	next.record(AgentCreated{AgentID: next.id, Name: next.name, Version: next.version, Occurred: f.clock.Now()})
	return next, nil
}
