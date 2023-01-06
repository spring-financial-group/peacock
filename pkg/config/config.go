package config

import "github.com/ilyakaznacheev/cleanenv"

type Config struct {
	LogLevel string `env:"LOG_LEVEL"`
	SCM      SCM    `yaml:"scm"`
}

type SCM struct {
	Provider string `yaml:"provider"`
	User     string `env:"GIT_USER"`
	Token    string `env:"GIT_TOKEN"`
	Secret   string `env:"HMAC_TOKEN"`
}

func Load() (*Config, error) {
	var cfg Config
	err := cleanenv.ReadConfig("config.yaml", &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
