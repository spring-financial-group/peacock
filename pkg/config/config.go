package config

import "github.com/ilyakaznacheev/cleanenv"

type Config struct {
	LogLevel        string `env:"LOG_LEVEL"`
	SCM             SCM
	MessageHandlers MessageHandlers
	DataSources     DataSources
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

func Load() (*Config, error) {
	var cfg Config
	err := cleanenv.ReadEnv(&cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
