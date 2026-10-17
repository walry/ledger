package tool

// ToolId 是工具聚合根的标识。
type ToolId string

// NewToolId 由原始字符串构造标识。
func NewToolId(v string) ToolId { return ToolId(v) }

// String 返回字符串形式。
func (id ToolId) String() string { return string(id) }

// IsZero 判断标识是否为空。
func (id ToolId) IsZero() bool { return id == "" }
