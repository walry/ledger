package modelconfig

import "errors"

// ProviderType 是模型供应商类型。
type ProviderType string

const (
	// ProviderOpenAI 标准 OpenAI。
	ProviderOpenAI ProviderType = "openai"
	// ProviderAnthropic Anthropic Claude。
	ProviderAnthropic ProviderType = "anthropic"
	// ProviderOpenAICompatible 兼容 OpenAI 协议的第三方服务。
	ProviderOpenAICompatible ProviderType = "openai_compatible"
)

// Valid 判断供应商类型是否为已知值。
func (p ProviderType) Valid() bool {
	switch p {
	case ProviderOpenAI, ProviderAnthropic, ProviderOpenAICompatible:
		return true
	default:
		return false
	}
}

// ModelParameters 是模型调用参数（值对象）。
type ModelParameters struct {
	Temperature      float64
	TopP             float64
	MaxTokens        int
	PresencePenalty  float64
	FrequencyPenalty float64
	Stop             []string
}

// DefaultParameters 返回一组合理默认值。
func DefaultParameters() ModelParameters {
	return ModelParameters{
		Temperature: 0.3,
		TopP:        1.0,
		MaxTokens:   2048,
	}
}

// Valid 校验参数范围。
func (p ModelParameters) Valid() error {
	if p.Temperature < 0 || p.Temperature > 2 {
		return errors.New("modelconfig: temperature must be within [0,2]")
	}
	if p.TopP < 0 || p.TopP > 1 {
		return errors.New("modelconfig: top_p must be within [0,1]")
	}
	if p.MaxTokens < 0 {
		return errors.New("modelconfig: max_tokens must be non-negative")
	}
	return nil
}
