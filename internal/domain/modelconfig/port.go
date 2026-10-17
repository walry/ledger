package modelconfig

import (
	"context"
	"encoding/json"
)

// Role 是对话角色。
type Role string

const (
	// RoleSystem 系统提示。
	RoleSystem Role = "system"
	// RoleUser 用户消息。
	RoleUser Role = "user"
	// RoleAssistant 模型回复。
	RoleAssistant Role = "assistant"
	// RoleTool 工具结果。
	RoleTool Role = "tool"
)

// ToolCall 是模型给出的工具调用。
type ToolCall struct {
	ID        string
	Name      string
	Arguments string // JSON 字符串
}

// Message 是发送给模型的一条消息（领域语言，不含任何 SDK 类型）。
type Message struct {
	Role       Role
	Content    string
	Name       string
	ToolCalls  []ToolCall
	ToolCallID string
}

// ToolSchema 是暴露给模型的工具定义。
type ToolSchema struct {
	Name        string
	Description string
	Parameters  json.RawMessage
}

// Request 是一次模型调用请求。
type Request struct {
	Model    string
	System   string
	Messages []Message
	Tools    []ToolSchema
	Params   ModelParameters
	Stream   bool
}

// Usage 是 token 使用统计。
type Usage struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

// Response 是一次非流式响应。
type Response struct {
	Content      string
	ToolCalls    []ToolCall
	FinishReason string
	Usage        Usage
}

// Chunk 是流式响应的增量。
type Chunk struct {
	Delta         string
	ToolCallDelta *ToolCall
	Done          bool
	Usage         *Usage
	Err           error
}

// ClientPort 是模型调用端口。
//
// 上层（runtime 领域服务）只依赖本接口，具体 SDK/HTTP 调用
// 位于 infrastructure/llm。
type ClientPort interface {
	Invoke(ctx context.Context, req Request) (Response, error)
	Stream(ctx context.Context, req Request) (<-chan Chunk, error)
}

// ClientFactory 依据配置创建客户端，并在配置变更时收到回调。
type ClientFactory interface {
	Create(cfg *ModelConfig) (ClientPort, error)
	OnConfigUpdated(ctx context.Context, cfg *ModelConfig) error
}
