package tool

import "errors"

// SourceKind 描述工具来源类型（值对象枚举）。
type SourceKind string

const (
	// SourceBuiltin 内置工具（进程内 Handler）。
	SourceBuiltin SourceKind = "builtin"
	// SourceHTTP 通过 HTTP 调用的外部工具。
	SourceHTTP SourceKind = "http"
	// SourceMCP 通过 MCP 协议接入的工具。
	SourceMCP SourceKind = "mcp"
)

// ToolSource 描述工具的实现来源（值对象）。
type ToolSource struct {
	kind   SourceKind
	uri    string
	config map[string]string
}

// NewToolSource 创建来源描述。
func NewToolSource(kind SourceKind, uri string, config map[string]string) (ToolSource, error) {
	switch kind {
	case SourceBuiltin, SourceHTTP, SourceMCP:
	default:
		return ToolSource{}, errors.New("tool: invalid source kind")
	}
	if kind != SourceBuiltin && uri == "" {
		return ToolSource{}, errors.New("tool: uri is required for non-builtin source")
	}
	if config == nil {
		config = map[string]string{}
	}
	cp := make(map[string]string, len(config))
	for k, v := range config {
		cp[k] = v
	}
	return ToolSource{kind: kind, uri: uri, config: cp}, nil
}

// Kind 返回来源类型。
func (s ToolSource) Kind() SourceKind { return s.kind }

// URI 返回来源地址。
func (s ToolSource) URI() string { return s.uri }

// Config 返回配置副本。
func (s ToolSource) Config() map[string]string {
	out := make(map[string]string, len(s.config))
	for k, v := range s.config {
		out[k] = v
	}
	return out
}
