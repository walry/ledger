package agent

import (
	"errors"
	"regexp"
	"strings"
)

var promptVarPattern = regexp.MustCompile(`\{\{\s*([a-zA-Z0-9_]+)\s*\}\}`)

// SystemPrompt 是系统提示词（值对象）。
type SystemPrompt struct {
	template  string
	variables []string
}

// NewSystemPrompt 创建系统提示词，并解析其中的 {{变量}}。
func NewSystemPrompt(template string) (SystemPrompt, error) {
	if strings.TrimSpace(template) == "" {
		return SystemPrompt{}, errors.New("agent: system prompt is required")
	}
	seen := map[string]struct{}{}
	vars := make([]string, 0)
	for _, m := range promptVarPattern.FindAllStringSubmatch(template, -1) {
		name := m[1]
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		vars = append(vars, name)
	}
	return SystemPrompt{template: template, variables: vars}, nil
}

// Template 返回原始模板。
func (p SystemPrompt) Template() string { return p.template }

// Variables 返回模板中的变量名。
func (p SystemPrompt) Variables() []string {
	out := make([]string, len(p.variables))
	copy(out, p.variables)
	return out
}

// Render 用给定变量渲染模板，未提供的变量替换为空串。
func (p SystemPrompt) Render(vars map[string]string) string {
	return promptVarPattern.ReplaceAllStringFunc(p.template, func(match string) string {
		name := promptVarPattern.FindStringSubmatch(match)[1]
		if v, ok := vars[name]; ok {
			return v
		}
		return ""
	})
}
