package feathers

import (
	"github.com/pkg/errors"
	"github.com/spring-financial-group/peacock/pkg/handlers"
	"github.com/spring-financial-group/peacock/pkg/utils"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"regexp"
)

const (
	feathersPath        = ".peacock/feathers.yaml"
	slackChannelIDRegex = "^[A-Z0-9]{9,11}$"
)

type Feathers struct {
	Teams []Team `yaml:"teams"`
}

type Team struct {
	Name        string   `yaml:"name"`
	ContactType string   `yaml:"contactType"`
	Addresses   []string `yaml:"addresses"`
}

func LoadFeathers() (*Feathers, error) {
	exists, err := utils.Exists(feathersPath)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.Errorf("could not find %s", feathersPath)
	}

	data, err := ioutil.ReadFile(feathersPath)
	if err != nil {
		return nil, err
	}

	feathers := new(Feathers)
	err = yaml.Unmarshal(data, feathers)
	if err != nil {
		return nil, err
	}
	return feathers, feathers.validate()
}

func (f *Feathers) validate() error {
	if f.Teams == nil {
		return errors.New("no teams found in feathers")
	}
	for _, team := range f.Teams {
		err := team.validate()
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *Feathers) GetTeamsByNames(name ...string) []Team {
	var teams []Team
	for _, tName := range name {
		for _, t := range f.Teams {
			if t.Name == tName {
				teams = append(teams, t)
			}
		}
	}
	return teams
}

func (f *Feathers) GetAllTeamNames() []string {
	var names []string
	for _, t := range f.Teams {
		names = append(names, t.Name)
	}
	return names
}

func (f *Feathers) GetAllContactTypes() []string {
	var types []string
	for _, t := range f.Teams {
		types = append(types, t.ContactType)
	}
	return types
}

func (f *Feathers) GetContactTypesByTeamNames(names ...string) []string {
	var types []string
	for _, t := range f.GetTeamsByNames(names...) {
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
		case handlers.Slack:
			match := slackRegex.MatchString(address)
			if !match {
				return errors.Errorf("failed to parse slack channel ID \"%s\" for team \"%s\"", address, t.Name)
			}
		}
	}
	return nil
}
