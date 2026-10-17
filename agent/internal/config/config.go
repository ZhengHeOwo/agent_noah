package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

const (
	defaultModelName    = "deepseek-flash"
	defaultEndpoint     = "https://api.deepseek.com/v1/chat/completions"
	defaultTimeout      = "60s"
	defaultSystemPrompt = "你是Bot,使用对方语言,回复上限50字符的内容,合理安排句式,冷静处理问题,只回答专业问题,拒绝回复日常问题"
)

type Config struct {
	Model        ModelConfig
	Agent        AgentConfig
	WebSearch    WebSearchConfig
	DatabasePath DatabasePath
}

type ModelConfig struct {
	APIKey   string
	Name     string
	Endpoint string
	Timeout  time.Duration
}

type AgentConfig struct {
	SystemPrompt string
}

type WebSearchConfig struct {
	BraveAPIKey string
}

type DatabasePath struct {
	SqlitePath string
}

func envOrDefault(key, defaultValue string) string {
	envValue := strings.TrimSpace(os.Getenv(key))
	if envValue == "" {
		return defaultValue
	}

	return envValue
}

func LoadConfig() (Config, error) {
	apiKey := strings.TrimSpace(os.Getenv("AI_API_KEY"))
	if apiKey == "" {
		return Config{}, fmt.Errorf("[环境变量] AI_API_KEY 未配置")
	}

	braveAPIKey := strings.TrimSpace(os.Getenv("BRAVE_API_KEY"))
	if braveAPIKey == "" {
		return Config{}, fmt.Errorf("[环境变量] BRAVE_API_KEY 未配置")
	}

	modelName := envOrDefault("MODEL", defaultModelName)
	modelEndpoint := envOrDefault("LLM_API_URL", defaultEndpoint)
	modelTimeoutText := envOrDefault("REQUEST_TIMEOUT", defaultTimeout)

	dataPrompt, err := os.ReadFile("./agent/local/prompts/mainModelPrompt.md")
	if err != nil {
		return Config{}, fmt.Errorf(
			"read file got analy_model_prompyt failed: %w",
			err,
		)
	}

	prompt := strings.TrimSpace(string(dataPrompt))
	if prompt == "" {
		return Config{}, fmt.Errorf("prompt must not be empty")
	}

	timeout, err := time.ParseDuration(modelTimeoutText)
	if err != nil {
		return Config{}, fmt.Errorf("[环境变量] REQUEST_TIMEOUT 解析失败: %w", err)
	}

	sqlitePath := envOrDefault("SQLITEPATH", "./agent/local/data/sqlite.db")

	config := Config{
		Model: ModelConfig{
			APIKey:   apiKey,
			Name:     modelName,
			Endpoint: modelEndpoint,
			Timeout:  timeout,
		},
		Agent: AgentConfig{
			SystemPrompt: prompt,
		},
		WebSearch: WebSearchConfig{
			BraveAPIKey: braveAPIKey,
		},
		DatabasePath: DatabasePath{
			SqlitePath: sqlitePath,
		},
	}

	return config, nil
}
