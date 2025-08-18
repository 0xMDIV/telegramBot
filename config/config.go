package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	BotToken string         `json:"bot_token"`
	Debug    bool           `json:"debug"`
	Captcha  CaptchaConfig  `json:"captcha"`
	Admin    AdminConfig    `json:"admin"`
	Database DatabaseConfig `json:"database"`
}

type CaptchaConfig struct {
	TimeoutMinutes int    `json:"timeout_minutes"`
	MaxAttempts    int    `json:"max_attempts"`
	WelcomeMessage string `json:"welcome_message"`
}

type AdminConfig struct {
	DefaultMuteHours   int `json:"default_mute_hours"`
	MaxDeleteMessages  int `json:"max_delete_messages"`
}

type DatabaseConfig struct {
	FilePath string `json:"file_path"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if config.BotToken == "" || config.BotToken == "YOUR_BOT_TOKEN_HERE" {
		return nil, fmt.Errorf("bot token not configured")
	}

	return &config, nil
}