package conversation

import "time"

// Role 是消息角色。
type Role string

const (
	// RoleSystem 系统消息。
	RoleSystem Role = "system"
	// RoleUser 用户消息。
	RoleUser Role = "user"
	// RoleAssistant 助手消息。
	RoleAssistant Role = "assistant"
	// RoleTool 工具结果消息。
	RoleTool Role = "tool"
)

// ToolCall 是助手给出的工具调用（值对象）。
type ToolCall struct {
	ID        string
	Name      string
	Arguments string
}

// Message 是会话消息（值对象）。
type Message struct {
	role       Role
	content    string
	name       string
	toolCalls  []ToolCall
	toolCallID string
	createdAt  time.Time
}

// NewUserMessage 创建用户消息。
func NewUserMessage(content string, now time.Time) Message {
	return Message{role: RoleUser, content: content, createdAt: now}
}

// NewSystemMessage 创建系统消息。
func NewSystemMessage(content string, now time.Time) Message {
	return Message{role: RoleSystem, content: content, createdAt: now}
}

// NewAssistantMessage 创建助手消息（可携带工具调用）。
func NewAssistantMessage(content string, toolCalls []ToolCall, now time.Time) Message {
	cp := make([]ToolCall, len(toolCalls))
	copy(cp, toolCalls)
	return Message{role: RoleAssistant, content: content, toolCalls: cp, createdAt: now}
}

// NewToolMessage 创建工具结果消息。
func NewToolMessage(toolCallID, name, content string, now time.Time) Message {
	return Message{role: RoleTool, content: content, name: name, toolCallID: toolCallID, createdAt: now}
}

// RehydrateMessage 从持久化状态重建消息。
func RehydrateMessage(role Role, content, name string, toolCalls []ToolCall, toolCallID string, createdAt time.Time) Message {
	return Message{role: role, content: content, name: name, toolCalls: toolCalls, toolCallID: toolCallID, createdAt: createdAt}
}

// Role 返回角色。
func (m Message) Role() Role { return m.role }

// Content 返回内容。
func (m Message) Content() string { return m.content }

// Name 返回工具名（仅 tool 消息）。
func (m Message) Name() string { return m.name }

// ToolCalls 返回工具调用副本。
func (m Message) ToolCalls() []ToolCall {
	out := make([]ToolCall, len(m.toolCalls))
	copy(out, m.toolCalls)
	return out
}

// ToolCallID 返回对应的工具调用 ID。
func (m Message) ToolCallID() string { return m.toolCallID }

// CreatedAt 返回创建时间。
func (m Message) CreatedAt() time.Time { return m.createdAt }
