package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/pkg/errors"
	"github.com/spring-financial-group/peacock/pkg/handlers"
	"github.com/spring-financial-group/peacock/pkg/utils"
	"net/mail"
	"regexp"
)

const (
	configPath          = ".peacock/config.yaml"
	slackChannelIDRegex = "^[A-Z0-9]{11}$"
)

type Config struct {
	Teams []Team `yaml:"teams"`
}

type Team struct {
	Name        string   `yaml:"name"`
	ContactType string   `yaml:"contactType"`
	Addresses   []string `yaml:"addresses"`
}

func LoadConfig() (*Config, error) {
	exists, err := utils.Exists(configPath)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.Errorf("could not find %s", configPath)
	}

	cfg := new(Config)
	err = cleanenv.ReadConfig(configPath, cfg)
	if err != nil {
		return nil, err
	}
	return cfg, cfg.validate()
}

func (f *Config) validate() error {
	if f.Teams == nil {
		return errors.New("no teams found in config")
	}
	for _, team := range f.Teams {
		err := team.validate()
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *Config) GetTeamByName(name string) *Team {
	for _, t := range f.Teams {
		if t.Name == name {
			return &t
		}
	}
	return nil
}

func (f *Config) GetAllTeamNames() []string {
	var names []string
	for _, t := range f.Teams {
		names = append(names, t.Name)
	}
	return names
}

func (f *Config) GetAllContactTypes() []string {
	var types []string
	for _, t := range f.Teams {
		types = append(types, t.ContactType)
	}
	return types
}

func (t *Team) validate() error {
	if t.Name == "" {
		return errors.New("no team name found")
	}
	if t.Addresses == nil {
		return errors.Errorf("team \"%s\" has no addresses", t.Name)
	}
	if t.ContactType == "" {
		return errors.Errorf("team \"%s\" has no contact type", t.Name)
	}

	// We should check that the contactType actually has a handler
	var valid bool
	for _, h := range handlers.Valid {
		if t.ContactType == h {
			valid = true
		}
	}
	if !valid {
		return errors.Errorf("team \"%s\" has an invalid contact type of \"%s\"", t.Name, t.ContactType)
	}

	// We should check that the addresses conform to the contact type
	slackRegex, err := regexp.Compile(slackChannelIDRegex)
	if err != nil {
		return err
	}
	for _, address := range t.Addresses {
		switch t.ContactType {
		case handlers.Email:
			_, err := mail.ParseAddress(address)
			if err != nil {
				return errors.Wrapf(err, "failed to parse email address \"%s\" for team \"%s\"", address, t.Name)
			}
		case handlers.Slack:
			match := slackRegex.MatchString(address)
			if !match {
				return errors.Errorf("failed to parse slack channel ID \"%s\" for team \"%s\"", address, t.Name)
			}
		}
	}
	return nil
}
