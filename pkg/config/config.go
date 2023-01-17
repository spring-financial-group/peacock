package config

import "github.com/ilyakaznacheev/cleanenv"

type Config struct {
	LogLevel        string          `env:"LOG_LEVEL"`
	SCM             SCM             `yaml:"scm"`
	MessageHandlers MessageHandlers `yaml:"messageHandlers"`
}

type SCM struct {
	User   string `env:"GIT_USER"`
	Token  string `env:"GIT_TOKEN"`
	Secret string `env:"GITHUB_SECRET"`
}

type MessageHandlers struct {
	Slack   Slack   `yaml:"slack"`
	Webhook Webhook `yaml:"webhook"`
}

type Slack struct {
	Token string `env:"SLACK_TOKEN"`
}

type Webhook struct {
	URL    string `env:"WEBHOOK_URL"`
	Token  string `env:"WEBHOOK_SECRET"`
	Secret string `env:"WEBHOOK_TOKEN"`
}

func Load() (*Config, error) {
	var cfg Config
	err := cleanenv.ReadConfig("config.yaml", &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
