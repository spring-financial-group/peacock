package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/spring-financial-group/peacock/pkg/utils"
	"os"
)

const (
	defaultPath = "development-config.yaml"
)

type Config struct {
	LogLevel        string `env:"LOG_LEVEL"`
	SCM             SCM
	MessageHandlers MessageHandlers
	DataSources     DataSources
	Cors            Cors `yaml:"cors"`
}

type DataSources struct {
	MongoDB struct {
		ConnectionString string `env:"MONGODB_CONNECTION_STRING"`
	}
}

type SCM struct {
	User   string `env:"GIT_USER"`
	Token  string `env:"GIT_TOKEN"`
	Secret string `env:"GITHUB_SECRET"`
}

type MessageHandlers struct {
	Slack   Slack
	Webhook Webhook
}

type Slack struct {
	Token string `env:"SLACK_TOKEN"`
}

type Webhook struct {
	URL    string `env:"WEBHOOK_URL"`
	Token  string `env:"WEBHOOK_TOKEN"`
	Secret string `env:"WEBHOOK_SECRET"`
}

type Cors struct {
	AllowOrigins    []string `yaml:"allowOrigins" env:"CORS_ALLOW_ORIGINS" envSeparator:","`
	AllowAllOrigins bool     `yaml:"allowAllOrigins" env:"CORS_ALLOW_ALL_ORIGINS"`
}

func Load() (*Config, error) {
	configPath := os.Getenv("CONFIG_PATH")
	exists, err := utils.Exists(configPath)
	if err != nil {
		return nil, err
	}
	if !exists {
		configPath = defaultPath
	}

	var cfg Config
	err = cleanenv.ReadConfig(configPath, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
