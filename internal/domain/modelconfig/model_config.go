package modelconfig

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/walry/ledger/internal/domain/ddd"
)

// ModelConfig 是模型连接配置聚合根（配置态）。
type ModelConfig struct {
	id        ModelConfigId
	name      string
	provider  ProviderType
	baseURL   string
	apiKey    string
	model     string
	params    ModelParameters
	enabled   bool
	createdAt time.Time
	updatedAt time.Time
	events    []ddd.DomainEvent
}

// NewModelConfig 创建模型配置。
func NewModelConfig(id ModelConfigId, name string, provider ProviderType, baseURL, apiKey, model string, params ModelParameters, now time.Time) (*ModelConfig, error) {
	if id.IsZero() {
		return nil, errors.New("modelconfig: id is required")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("modelconfig: name is required")
	}
	if !provider.Valid() {
		return nil, errors.New("modelconfig: invalid provider")
	}
	model = strings.TrimSpace(model)
	if model == "" {
		return nil, errors.New("modelconfig: model is required")
	}
	if err := params.Valid(); err != nil {
		return nil, err
	}
	mc := &ModelConfig{
		id:        id,
		name:      name,
		provider:  provider,
		baseURL:   strings.TrimRight(baseURL, "/"),
		apiKey:    apiKey,
		model:     model,
		params:    params,
		enabled:   true,
		createdAt: now,
		updatedAt: now,
	}
	mc.record(ModelConfigCreated{ConfigID: id, Name: name, Provider: provider, Occurred: now})
	return mc, nil
}

// RehydrateModelConfig 从持久化状态重建配置。
func RehydrateModelConfig(
	id ModelConfigId, name string, provider ProviderType, baseURL, apiKey, model string,
	params ModelParameters, enabled bool, createdAt, updatedAt time.Time,
) *ModelConfig {
	return &ModelConfig{
		id: id, name: name, provider: provider, baseURL: baseURL, apiKey: apiKey,
		model: model, params: params, enabled: enabled, createdAt: createdAt, updatedAt: updatedAt,
	}
}

// ID 返回聚合标识。
func (m *ModelConfig) ID() ModelConfigId { return m.id }

// Name 返回名称。
func (m *ModelConfig) Name() string { return m.name }

// Provider 返回供应商类型。
func (m *ModelConfig) Provider() ProviderType { return m.provider }

// BaseURL 返回服务地址。
func (m *ModelConfig) BaseURL() string { return m.baseURL }

// APIKey 返回密钥（仅基础设施层构造客户端时使用）。
func (m *ModelConfig) APIKey() string { return m.apiKey }

// Model 返回模型名。
func (m *ModelConfig) Model() string { return m.model }

// Parameters 返回调用参数。
func (m *ModelConfig) Parameters() ModelParameters { return m.params }

// Enabled 返回是否启用。
func (m *ModelConfig) Enabled() bool { return m.enabled }

// CreatedAt 返回创建时间。
func (m *ModelConfig) CreatedAt() time.Time { return m.createdAt }

// UpdatedAt 返回更新时间。
func (m *ModelConfig) UpdatedAt() time.Time { return m.updatedAt }

// Update 更新连接与参数信息并记录事件。
func (m *ModelConfig) Update(baseURL, apiKey, model string, params ModelParameters, now time.Time) error {
	if err := params.Valid(); err != nil {
		return fmt.Errorf("modelconfig: invalid parameters: %w", err)
	}
	if strings.TrimSpace(model) == "" {
		return errors.New("modelconfig: model is required")
	}
	m.baseURL = strings.TrimRight(baseURL, "/")
	m.apiKey = apiKey
	m.model = model
	m.params = params
	m.updatedAt = now
	m.record(ModelConfigUpdated{ConfigID: m.id, Occurred: now})
	return nil
}

// Enable 启用配置。
func (m *ModelConfig) Enable(now time.Time) {
	if m.enabled {
		return
	}
	m.enabled = true
	m.updatedAt = now
	m.record(ModelConfigEnabledChanged{ConfigID: m.id, Enabled: true, Occurred: now})
}

// Disable 停用配置。
func (m *ModelConfig) Disable(now time.Time) {
	if !m.enabled {
		return
	}
	m.enabled = false
	m.updatedAt = now
	m.record(ModelConfigEnabledChanged{ConfigID: m.id, Enabled: false, Occurred: now})
}

// PullEvents 返回并清空领域事件。
func (m *ModelConfig) PullEvents() []ddd.DomainEvent {
	events := m.events
	m.events = nil
	return events
}

func (m *ModelConfig) record(e ddd.DomainEvent) { m.events = append(m.events, e) }
