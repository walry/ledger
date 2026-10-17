// Package config 从环境变量加载应用配置。
//
// 本包不依赖任何外部配置库；所有键名与默认值见 configs/config.example.yaml。
package config

import (
	"os"
	"strconv"
)

// Config 汇总应用的运行时配置。
type Config struct {
	HTTPAddr           string
	DatabaseDSN        string
	OpenAIAPIKey       string
	OpenAIBaseURL      string
	AnthropicAPIKey    string
	AnthropicBaseURL   string
	EmbeddingModel     string
	EmbeddingDimension int
	VectorDimension    int
	WebSearchEndpoint  string
	WebSearchAPIKey    string
	SQLDSN             string
	DefaultModel       string
	DefaultProvider    string
	AnalysisWorkers    int
	AutoMigrate        bool
}

// Load 读取环境变量并填充默认值。
func Load() *Config {
	return &Config{
		HTTPAddr:         getString("HTTP_ADDR", ":8080"),
		DatabaseDSN:      getString("DATABASE_DSN", ""),
		OpenAIAPIKey:     getString("OPENAI_API_KEY", ""),
		OpenAIBaseURL:    getString("OPENAI_BASE_URL", "https://api.openai.com/v1"),
		AnthropicAPIKey:  getString("ANTHROPIC_API_KEY", ""),
		AnthropicBaseURL: getString("ANTHROPIC_BASE_URL", "https://api.anthropic.com"),

		EmbeddingModel:     getString("EMBEDDING_MODEL", ""),
		EmbeddingDimension: getInt("EMBEDDING_DIMENSION", 1536),
		VectorDimension:    getInt("VECTOR_DIMENSION", 1536),

		WebSearchEndpoint: getString("WEB_SEARCH_ENDPOINT", ""),
		WebSearchAPIKey:   getString("WEB_SEARCH_API_KEY", ""),
		SQLDSN:            getString("SQL_DSN", ""),

		DefaultModel:    getString("DEFAULT_MODEL", ""),
		DefaultProvider: getString("DEFAULT_PROVIDER", "openai"),
		AnalysisWorkers: getInt("ANALYSIS_WORKERS", 2),
		AutoMigrate:     getBool("AUTO_MIGRATE", true),
	}
}

func getString(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

func getInt(key string, def int) int {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func getBool(key string, def bool) bool {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return def
}
