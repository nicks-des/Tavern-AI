package config

import (
	"os"
	"path/filepath"
)

type Config struct {
	DataDir   string
	HTTPPort  string
	OpenAIKey string
	OpenAIURL string
	ModelName string
}

func Load() *Config {
	dataDir := envOrDefault("TAVERN_DATA_DIR", filepath.Join(".", "data"))
	return &Config{
		DataDir:   dataDir,
		HTTPPort:  envOrDefault("TAVERN_HTTP_PORT", "8081"),
		OpenAIKey: os.Getenv("OPENAI_API_KEY"),
		OpenAIURL: envOrDefault("OPENAI_BASE_URL", "https://api.openai.com/v1"),
		ModelName: envOrDefault("LLM_MODEL", "deepseek-chat"),
	}
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
