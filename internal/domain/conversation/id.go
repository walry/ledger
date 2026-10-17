package conversation

// ConversationId 是会话聚合根的标识。
type ConversationId string

// NewConversationId 由原始字符串构造标识。
func NewConversationId(v string) ConversationId { return ConversationId(v) }

// String 返回字符串形式。
func (id ConversationId) String() string { return string(id) }

// IsZero 判断标识是否为空。
func (id ConversationId) IsZero() bool { return id == "" }

// TurnId 是轮次实体的标识。
type TurnId string

// NewTurnId 由原始字符串构造标识。
func NewTurnId(v string) TurnId { return TurnId(v) }

// String 返回字符串形式。
func (id TurnId) String() string { return string(id) }

// IsZero 判断标识是否为空。
func (id TurnId) IsZero() bool { return id == "" }
