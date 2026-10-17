package agent

// AgentId 是智能体聚合根的标识。
type AgentId string

// NewAgentId 由原始字符串构造标识。
func NewAgentId(v string) AgentId { return AgentId(v) }

// String 返回字符串形式。
func (id AgentId) String() string { return string(id) }

// IsZero 判断标识是否为空。
func (id AgentId) IsZero() bool { return id == "" }
