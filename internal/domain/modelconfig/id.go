package modelconfig

// ModelConfigId 是模型配置聚合根的标识。
type ModelConfigId string

// NewModelConfigId 由原始字符串构造标识。
func NewModelConfigId(v string) ModelConfigId { return ModelConfigId(v) }

// String 返回字符串形式。
func (id ModelConfigId) String() string { return string(id) }

// IsZero 判断标识是否为空。
func (id ModelConfigId) IsZero() bool { return id == "" }
