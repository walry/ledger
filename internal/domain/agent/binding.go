package agent

import (
	"errors"

	"github.com/walry/ledger/internal/domain/modelconfig"
	"github.com/walry/ledger/internal/domain/tool"
)

// ModelConfigRef 是智能体对模型配置的引用（值对象，跨聚合只引用 ID）。
type ModelConfigRef struct {
	ConfigID modelconfig.ModelConfigId
	Model    string
	Provider modelconfig.ProviderType
}

// NewModelConfigRef 创建模型配置引用。
func NewModelConfigRef(configID modelconfig.ModelConfigId, model string, provider modelconfig.ProviderType) (ModelConfigRef, error) {
	if configID.IsZero() {
		return ModelConfigRef{}, errors.New("agent: model config id is required")
	}
	if model == "" {
		return ModelConfigRef{}, errors.New("agent: model name is required")
	}
	return ModelConfigRef{ConfigID: configID, Model: model, Provider: provider}, nil
}

// IsZero 判断引用是否为空。
func (r ModelConfigRef) IsZero() bool { return r.ConfigID.IsZero() }

// SkillBinding 是技能绑定（技能目录不在本应用内建模，按 ID/版本引用）。
type SkillBinding struct {
	SkillID string
	Version int
	Enabled bool
}

// NewSkillBinding 创建技能绑定。
func NewSkillBinding(skillID string, version int) (SkillBinding, error) {
	if skillID == "" {
		return SkillBinding{}, errors.New("agent: skill id is required")
	}
	if version <= 0 {
		version = 1
	}
	return SkillBinding{SkillID: skillID, Version: version, Enabled: true}, nil
}

// ToolBinding 是工具绑定（含 ID、版本与配置）。
type ToolBinding struct {
	ToolID  tool.ToolId
	Version int
	Enabled bool
	Config  map[string]string
}

// NewToolBinding 创建工具绑定。
func NewToolBinding(toolID tool.ToolId, version int, config map[string]string) (ToolBinding, error) {
	if toolID.IsZero() {
		return ToolBinding{}, errors.New("agent: tool id is required")
	}
	if version <= 0 {
		version = 1
	}
	if config == nil {
		config = map[string]string{}
	}
	cp := make(map[string]string, len(config))
	for k, v := range config {
		cp[k] = v
	}
	return ToolBinding{ToolID: toolID, Version: version, Enabled: true, Config: cp}, nil
}
