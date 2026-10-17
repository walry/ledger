package llm

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/walry/ledger/internal/domain/modelconfig"
)

// 供应商默认服务地址。
const (
	defaultOpenAIBaseURL    = "https://api.openai.com/v1"
	defaultAnthropicBaseURL = "https://api.anthropic.com/v1"
)

// cachedClient 保存某配置在某版本下的客户端实例。
type cachedClient struct {
	fingerprint string
	client      modelconfig.ClientPort
}

// ClientFactory 依据 ModelConfig 创建并缓存模型客户端。
// 缓存键为配置 ID，值为客户端 + 版本指纹；配置变更时失效重建。
type ClientFactory struct {
	mu      sync.Mutex
	clients map[modelconfig.ModelConfigId]cachedClient
}

// NewClientFactory 构造客户端工厂。
func NewClientFactory() *ClientFactory {
	return &ClientFactory{clients: make(map[modelconfig.ModelConfigId]cachedClient)}
}

var _ modelconfig.ClientFactory = (*ClientFactory)(nil)

// Create 按配置返回客户端，命中缓存且版本一致时复用。
func (f *ClientFactory) Create(cfg *modelconfig.ModelConfig) (modelconfig.ClientPort, error) {
	if cfg == nil {
		return nil, errors.New("llm: model config is nil")
	}
	if strings.TrimSpace(cfg.APIKey()) == "" {
		return nil, fmt.Errorf("llm: api key is required for model config %q", cfg.ID().String())
	}

	fp := fingerprint(cfg)

	f.mu.Lock()
	defer f.mu.Unlock()

	if cached, ok := f.clients[cfg.ID()]; ok && cached.fingerprint == fp {
		return cached.client, nil
	}
	client, err := build(cfg)
	if err != nil {
		return nil, err
	}
	f.clients[cfg.ID()] = cachedClient{fingerprint: fp, client: client}
	return client, nil
}

// OnConfigUpdated 使该配置的缓存客户端失效，下次 Create 时重建。
func (f *ClientFactory) OnConfigUpdated(_ context.Context, cfg *modelconfig.ModelConfig) error {
	if cfg == nil {
		return nil
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.clients, cfg.ID())
	return nil
}

// build 依据供应商类型构造具体客户端，空 baseURL 使用默认地址。
func build(cfg *modelconfig.ModelConfig) (modelconfig.ClientPort, error) {
	baseURL := cfg.BaseURL()
	switch cfg.Provider() {
	case modelconfig.ProviderOpenAI, modelconfig.ProviderOpenAICompatible:
		if baseURL == "" {
			baseURL = defaultOpenAIBaseURL
		}
		return NewOpenAIClient(baseURL, cfg.APIKey(), cfg.Model()), nil
	case modelconfig.ProviderAnthropic:
		if baseURL == "" {
			baseURL = defaultAnthropicBaseURL
		}
		return NewAnthropicClient(baseURL, cfg.APIKey(), cfg.Model()), nil
	default:
		return nil, fmt.Errorf("llm: unsupported provider %q", cfg.Provider())
	}
}

// fingerprint 计算连接与参数信息的版本指纹，用于缓存失效判断。
func fingerprint(cfg *modelconfig.ModelConfig) string {
	h := sha256.New()
	write := func(v string) {
		h.Write([]byte(v))
		h.Write([]byte{0})
	}
	write(string(cfg.Provider()))
	write(cfg.BaseURL())
	write(cfg.APIKey())
	write(cfg.Model())

	p := cfg.Parameters()
	write(strconv.FormatFloat(p.Temperature, 'g', -1, 64))
	write(strconv.FormatFloat(p.TopP, 'g', -1, 64))
	write(strconv.Itoa(p.MaxTokens))
	write(strconv.FormatFloat(p.PresencePenalty, 'g', -1, 64))
	write(strconv.FormatFloat(p.FrequencyPenalty, 'g', -1, 64))
	write(strings.Join(p.Stop, "\x1f"))
	return hex.EncodeToString(h.Sum(nil))
}
