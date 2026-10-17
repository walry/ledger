// Package agent provides the GORM persistence adapter for the agent bounded
// context. It implements the domain port agent.Repository.
package agent

import (
	"encoding/json"
	"time"

	domain "github.com/walry/ledger/internal/domain/agent"
	"github.com/walry/ledger/internal/domain/knowledge"
	"github.com/walry/ledger/internal/domain/modelconfig"
)

// AgentRow is the persistence model for the Agent aggregate (configuration
// state). Its value-object collections are stored as JSONB columns.
type AgentRow struct {
	ID           string     `gorm:"column:id;primaryKey;type:text"`
	Name         string     `gorm:"column:name;type:text;not null;index"`
	Description  string     `gorm:"column:description;type:text"`
	Status       string     `gorm:"column:status;type:text;not null;index"`
	SystemPrompt string     `gorm:"column:system_prompt;type:text"`
	ModelConfig  string     `gorm:"column:model_config_id;type:text"`
	Model        string     `gorm:"column:model;type:text"`
	Provider     string     `gorm:"column:provider;type:text"`
	Skills       []byte     `gorm:"column:skills;type:jsonb"`
	Tools        []byte     `gorm:"column:tools;type:jsonb"`
	KBIDs        []byte     `gorm:"column:kb_ids;type:jsonb"`
	MaxSteps     int        `gorm:"column:max_steps;not null;default:6"`
	RetrievalTop int        `gorm:"column:retrieval_top_k;not null;default:5"`
	Version      int        `gorm:"column:version;not null;default:1"`
	ParentVer    int        `gorm:"column:parent_version;not null;default:0"`
	CreatedAt    time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt    time.Time  `gorm:"column:updated_at;not null"`
	PublishedAt  *time.Time `gorm:"column:published_at"`
}

// TableName pins the physical table name.
func (AgentRow) TableName() string { return "agents" }

func marshalJSON(v any, fallback string) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		return []byte(fallback)
	}
	return b
}

func unmarshalSkills(b []byte) []domain.SkillBinding {
	out := make([]domain.SkillBinding, 0)
	if len(b) == 0 {
		return out
	}
	if err := json.Unmarshal(b, &out); err != nil {
		return make([]domain.SkillBinding, 0)
	}
	if out == nil {
		out = make([]domain.SkillBinding, 0)
	}
	return out
}

func unmarshalTools(b []byte) []domain.ToolBinding {
	out := make([]domain.ToolBinding, 0)
	if len(b) == 0 {
		return out
	}
	if err := json.Unmarshal(b, &out); err != nil {
		return make([]domain.ToolBinding, 0)
	}
	if out == nil {
		out = make([]domain.ToolBinding, 0)
	}
	for i := range out {
		if out[i].Config == nil {
			out[i].Config = map[string]string{}
		}
	}
	return out
}

func unmarshalKBIDs(b []byte) []knowledge.KnowledgeBaseId {
	out := make([]knowledge.KnowledgeBaseId, 0)
	if len(b) == 0 {
		return out
	}
	if err := json.Unmarshal(b, &out); err != nil {
		return make([]knowledge.KnowledgeBaseId, 0)
	}
	if out == nil {
		out = make([]knowledge.KnowledgeBaseId, 0)
	}
	return out
}

func agentToRow(a *domain.Agent) *AgentRow {
	ref := a.ModelRef()
	ctx := a.Context()
	return &AgentRow{
		ID:           a.ID().String(),
		Name:         a.Name(),
		Description:  a.Description(),
		Status:       string(a.Status()),
		SystemPrompt: a.SystemPrompt().Template(),
		ModelConfig:  ref.ConfigID.String(),
		Model:        ref.Model,
		Provider:     string(ref.Provider),
		Skills:       marshalJSON(a.Skills(), "[]"),
		Tools:        marshalJSON(a.Tools(), "[]"),
		KBIDs:        marshalJSON(ctx.KnowledgeBaseIDs(), "[]"),
		MaxSteps:     ctx.MaxSteps(),
		RetrievalTop: ctx.RetrievalTopK(),
		Version:      a.Version(),
		ParentVer:    a.ParentVersion(),
		CreatedAt:    a.CreatedAt(),
		UpdatedAt:    a.UpdatedAt(),
		PublishedAt:  a.PublishedAt(),
	}
}

func agentFromRow(row *AgentRow) *domain.Agent {
	prompt, err := domain.NewSystemPrompt(row.SystemPrompt)
	if err != nil {
		// A persisted agent must always carry a non-empty prompt; fall back to a
		// non-whitespace placeholder so rehydration never fails.
		prompt, _ = domain.NewSystemPrompt("-")
	}

	ref, err := domain.NewModelConfigRef(
		modelconfig.ModelConfigId(row.ModelConfig),
		row.Model,
		modelconfig.ProviderType(row.Provider),
	)
	if err != nil {
		ref = domain.ModelConfigRef{}
	}

	ctx, err := domain.NewAgentContext(unmarshalKBIDs(row.KBIDs), row.MaxSteps, row.RetrievalTop)
	if err != nil {
		ctx = domain.DefaultAgentContext()
	}

	return domain.RehydrateAgent(
		domain.AgentId(row.ID), row.Name, row.Description, domain.Status(row.Status),
		prompt, ref, unmarshalSkills(row.Skills), unmarshalTools(row.Tools), ctx,
		row.Version, row.ParentVer, row.CreatedAt, row.UpdatedAt, row.PublishedAt,
	)
}
