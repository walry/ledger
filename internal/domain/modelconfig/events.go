package modelconfig

import (
	"time"

	"github.com/walry/ledger/internal/domain/ddd"
)

// ModelConfigCreated 模型配置已创建。
type ModelConfigCreated struct {
	ConfigID ModelConfigId
	Name     string
	Provider ProviderType
	Occurred time.Time
}

// EventName 实现 ddd.DomainEvent。
func (ModelConfigCreated) EventName() string { return "ModelConfigCreated" }

// OccurredAt 实现 ddd.DomainEvent。
func (e ModelConfigCreated) OccurredAt() time.Time { return e.Occurred }

// ModelConfigUpdated 模型配置已更新。llm 工厂据此刷新客户端缓存。
type ModelConfigUpdated struct {
	ConfigID ModelConfigId
	Occurred time.Time
}

// EventName 实现 ddd.DomainEvent。
func (ModelConfigUpdated) EventName() string { return "ModelConfigUpdated" }

// OccurredAt 实现 ddd.DomainEvent。
func (e ModelConfigUpdated) OccurredAt() time.Time { return e.Occurred }

// ModelConfigEnabledChanged 启用状态变化。
type ModelConfigEnabledChanged struct {
	ConfigID ModelConfigId
	Enabled  bool
	Occurred time.Time
}

// EventName 实现 ddd.DomainEvent。
func (ModelConfigEnabledChanged) EventName() string { return "ModelConfigEnabledChanged" }

// OccurredAt 实现 ddd.DomainEvent。
func (e ModelConfigEnabledChanged) OccurredAt() time.Time { return e.Occurred }

var _ ddd.DomainEvent = ModelConfigCreated{}
var _ ddd.DomainEvent = ModelConfigUpdated{}
var _ ddd.DomainEvent = ModelConfigEnabledChanged{}
