package agent

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/walry/ledger/internal/domain/ddd"
	"github.com/walry/ledger/internal/domain/knowledge"
	"github.com/walry/ledger/internal/domain/tool"
)

// Agent 是智能体聚合根（配置态）。
//
// 不变量：仅草稿可编辑；发布后配置冻结，需通过新版本修改。
type Agent struct {
	id            AgentId
	name          string
	description   string
	status        Status
	systemPrompt  SystemPrompt
	modelRef      ModelConfigRef
	skills        []SkillBinding
	tools         []ToolBinding
	context       AgentContext
	version       int
	parentVersion int
	createdAt     time.Time
	updatedAt     time.Time
	publishedAt   *time.Time
	events        []ddd.DomainEvent
}

// NewAgent 创建草稿智能体。
func NewAgent(id AgentId, name, description string, prompt SystemPrompt, modelRef ModelConfigRef, ctx AgentContext, now time.Time) (*Agent, error) {
	if id.IsZero() {
		return nil, errors.New("agent: id is required")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("agent: name is required")
	}
	if modelRef.IsZero() {
		return nil, errors.New("agent: model config ref is required")
	}
	a := &Agent{
		id:           id,
		name:         name,
		description:  description,
		status:       StatusDraft,
		systemPrompt: prompt,
		modelRef:     modelRef,
		skills:       make([]SkillBinding, 0),
		tools:        make([]ToolBinding, 0),
		context:      ctx,
		version:      1,
		createdAt:    now,
		updatedAt:    now,
	}
	a.record(AgentCreated{AgentID: id, Name: name, Version: 1, Occurred: now})
	return a, nil
}

// RehydrateAgent 从持久化状态重建智能体。
func RehydrateAgent(
	id AgentId, name, description string, status Status, prompt SystemPrompt, modelRef ModelConfigRef,
	skills []SkillBinding, tools []ToolBinding, ctx AgentContext, version, parentVersion int,
	createdAt, updatedAt time.Time, publishedAt *time.Time,
) *Agent {
	if skills == nil {
		skills = make([]SkillBinding, 0)
	}
	if tools == nil {
		tools = make([]ToolBinding, 0)
	}
	return &Agent{
		id: id, name: name, description: description, status: status, systemPrompt: prompt,
		modelRef: modelRef, skills: skills, tools: tools, context: ctx, version: version,
		parentVersion: parentVersion, createdAt: createdAt, updatedAt: updatedAt, publishedAt: publishedAt,
	}
}

// ID 返回聚合标识。
func (a *Agent) ID() AgentId { return a.id }

// Name 返回名称。
func (a *Agent) Name() string { return a.name }

// Description 返回描述。
func (a *Agent) Description() string { return a.description }

// Status 返回状态。
func (a *Agent) Status() Status { return a.status }

// SystemPrompt 返回系统提示词。
func (a *Agent) SystemPrompt() SystemPrompt { return a.systemPrompt }

// ModelRef 返回模型配置引用。
func (a *Agent) ModelRef() ModelConfigRef { return a.modelRef }

// Skills 返回技能绑定副本。
func (a *Agent) Skills() []SkillBinding {
	out := make([]SkillBinding, len(a.skills))
	copy(out, a.skills)
	return out
}

// Tools 返回工具绑定副本。
func (a *Agent) Tools() []ToolBinding {
	out := make([]ToolBinding, len(a.tools))
	copy(out, a.tools)
	return out
}

// Context 返回配置态上下文。
func (a *Agent) Context() AgentContext { return a.context }

// Version 返回版本号。
func (a *Agent) Version() int { return a.version }

// ParentVersion 返回上一版本号。
func (a *Agent) ParentVersion() int { return a.parentVersion }

// CreatedAt 返回创建时间。
func (a *Agent) CreatedAt() time.Time { return a.createdAt }

// UpdatedAt 返回更新时间。
func (a *Agent) UpdatedAt() time.Time { return a.updatedAt }

// PublishedAt 返回发布时间。
func (a *Agent) PublishedAt() *time.Time { return a.publishedAt }

// UpdateSystemPrompt 更新系统提示词（仅草稿）。
func (a *Agent) UpdateSystemPrompt(prompt SystemPrompt, now time.Time) error {
	if err := a.ensureEditable(); err != nil {
		return err
	}
	a.systemPrompt = prompt
	a.touch(now)
	return nil
}

// UpdateModelRef 更新模型引用（仅草稿）。
func (a *Agent) UpdateModelRef(ref ModelConfigRef, now time.Time) error {
	if err := a.ensureEditable(); err != nil {
		return err
	}
	if ref.IsZero() {
		return errors.New("agent: model config ref is required")
	}
	a.modelRef = ref
	a.touch(now)
	return nil
}

// UpdateContext 更新上下文配置（仅草稿）。
func (a *Agent) UpdateContext(ctx AgentContext, now time.Time) error {
	if err := a.ensureEditable(); err != nil {
		return err
	}
	a.context = ctx
	a.touch(now)
	return nil
}

// AddSkill 绑定技能（幂等，仅草稿）。
func (a *Agent) AddSkill(b SkillBinding, now time.Time) error {
	if err := a.ensureEditable(); err != nil {
		return err
	}
	for i, existing := range a.skills {
		if existing.SkillID == b.SkillID {
			a.skills[i] = b
			a.touch(now)
			return nil
		}
	}
	a.skills = append(a.skills, b)
	a.touch(now)
	return nil
}

// RemoveSkill 解绑技能（仅草稿）。
func (a *Agent) RemoveSkill(skillID string, now time.Time) error {
	if err := a.ensureEditable(); err != nil {
		return err
	}
	for i, existing := range a.skills {
		if existing.SkillID == skillID {
			a.skills = append(a.skills[:i], a.skills[i+1:]...)
			a.touch(now)
			return nil
		}
	}
	return fmt.Errorf("agent: skill %s not bound", skillID)
}

// BindTool 绑定工具（幂等，仅草稿）。
func (a *Agent) BindTool(b ToolBinding, now time.Time) error {
	if err := a.ensureEditable(); err != nil {
		return err
	}
	for i, existing := range a.tools {
		if existing.ToolID == b.ToolID {
			a.tools[i] = b
			a.touch(now)
			return nil
		}
	}
	a.tools = append(a.tools, b)
	a.touch(now)
	return nil
}

// UnbindTool 解绑工具（仅草稿）。
func (a *Agent) UnbindTool(id tool.ToolId, now time.Time) error {
	if err := a.ensureEditable(); err != nil {
		return err
	}
	for i, existing := range a.tools {
		if existing.ToolID == id {
			a.tools = append(a.tools[:i], a.tools[i+1:]...)
			a.touch(now)
			return nil
		}
	}
	return fmt.Errorf("agent: tool %s not bound", id)
}

// EnabledToolIDs 返回已启用工具 ID。
func (a *Agent) EnabledToolIDs() []tool.ToolId {
	ids := make([]tool.ToolId, 0, len(a.tools))
	for _, b := range a.tools {
		if b.Enabled {
			ids = append(ids, b.ToolID)
		}
	}
	return ids
}

// KnowledgeBaseIDs 返回绑定的知识库 ID。
func (a *Agent) KnowledgeBaseIDs() []knowledge.KnowledgeBaseId {
	return a.context.KnowledgeBaseIDs()
}

// Publish 发布智能体（草稿→已发布）。
func (a *Agent) Publish(now time.Time) error {
	if a.status != StatusDraft {
		return fmt.Errorf("agent: only draft can be published, current=%s", a.status)
	}
	if a.systemPrompt.Template() == "" {
		return errors.New("agent: system prompt is required before publish")
	}
	a.status = StatusPublished
	a.publishedAt = &now
	a.touch(now)
	a.record(AgentPublished{AgentID: a.id, Version: a.version, Occurred: now})
	return nil
}

// Archive 归档智能体。
func (a *Agent) Archive(now time.Time) error {
	if a.status == StatusArchived {
		return nil
	}
	a.status = StatusArchived
	a.touch(now)
	a.record(AgentArchived{AgentID: a.id, Version: a.version, Occurred: now})
	return nil
}

// PullEvents 返回并清空领域事件。
func (a *Agent) PullEvents() []ddd.DomainEvent {
	events := a.events
	a.events = nil
	return events
}

func (a *Agent) ensureEditable() error {
	if !a.status.Editable() {
		return fmt.Errorf("agent: %s is not editable in status %s", a.id, a.status)
	}
	return nil
}

func (a *Agent) touch(now time.Time) { a.updatedAt = now }

func (a *Agent) record(e ddd.DomainEvent) { a.events = append(a.events, e) }
