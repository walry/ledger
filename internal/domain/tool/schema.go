package tool

import "encoding/json"

// JSONSchema 是工具参数的 JSON Schema（值对象）。
type JSONSchema struct {
	Type        string                `json:"type,omitempty"`
	Description string                `json:"description,omitempty"`
	Properties  map[string]JSONSchema `json:"properties,omitempty"`
	Required    []string              `json:"required,omitempty"`
	Items       *JSONSchema           `json:"items,omitempty"`
	Enum        []string              `json:"enum,omitempty"`
}

// Valid 判断 Schema 是否可用。
func (s JSONSchema) Valid() bool { return s.Type != "" }

// Raw 返回 Schema 的 JSON 表示。
func (s JSONSchema) Raw() json.RawMessage {
	b, err := json.Marshal(s)
	if err != nil {
		return json.RawMessage(`{"type":"object"}`)
	}
	return b
}

// ObjectSchema 构造一个最简单的 object Schema。
func ObjectSchema(properties map[string]JSONSchema, required ...string) JSONSchema {
	if properties == nil {
		properties = map[string]JSONSchema{}
	}
	return JSONSchema{Type: "object", Properties: properties, Required: required}
}
