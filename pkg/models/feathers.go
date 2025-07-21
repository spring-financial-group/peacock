package models

type Feathers struct {
	Teams  Teams  `yaml:"teams"`
	Config Config `yaml:"config"`
}

type Config struct {
	Messages           Messages `yaml:"messages"`
	PRBodyTemplatePath string   `yaml:"prBodyTemplatePath"`
}

type Messages struct {
	Subject string `yaml:"subject"`
}
