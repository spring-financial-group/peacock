package feathers

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/spring-financial-group/peacock/pkg/models"
	"github.com/spring-financial-group/peacock/pkg/utils"
	"gopkg.in/yaml.v3"
	"os"
	"regexp"
)

const (
	feathersPath        = ".peacock/feathers.yaml"
	slackChannelIDRegex = "^[A-Z0-9]{9,11}$"
)

type Feathers struct {
	Teams  []Team `yaml:"teams"`
	Config Config `yaml:"config"`
}

type Team struct {
	Name        string   `yaml:"name"`
	APIKey      string   `yaml:"apiKey"`
	ContactType string   `yaml:"contactType"`
	Addresses   []string `yaml:"addresses"`
}

type Config struct {
	Messages Messages `yaml:"messages"`
}

type Messages struct {
	Subject string `yaml:"subject"`
}

func GetFeathersFromFile() (*Feathers, error) {
	exists, err := utils.Exists(feathersPath)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.Errorf("could not find %s", feathersPath)
	}

	data, err := os.ReadFile(feathersPath)
	if err != nil {
		return nil, err
	}

	return GetFeathersFromBytes(data)
}

func GetFeathersFromBytes(data []byte) (*Feathers, error) {
	feathers := new(Feathers)
	err := yaml.Unmarshal(data, &feathers)
	if err != nil {
		return nil, err
	}
	return feathers, feathers.validate()
}

func (f *Feathers) validate() error {
	if f.Teams == nil {
		return errors.New("no teams found in feathers")
	}

	unique := map[string]map[string]bool{
		"Name":   make(map[string]bool),
		"APIKey": make(map[string]bool),
	}

	for _, team := range f.Teams {
		// Check that the individual teams are set up correctly
		err := team.validate()
		if err != nil {
			return err
		}

		// Check the team names and API keys are unique
		if _, exists := unique["Name"][team.Name]; exists {
			return fmt.Errorf("duplicate team name found: %s", team.Name)
		}
		if _, exists := unique["APIKey"][team.APIKey]; exists {
			return fmt.Errorf("duplicate apiKey found: %s", team.APIKey)
		}
		unique["Name"][team.Name] = true
		unique["APIKey"][team.APIKey] = true
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

func (f *Feathers) ExistsInFeathers(teamNames ...string) error {
	allTeamsInFeathers := f.GetAllTeamNames()
	for _, name := range teamNames {
		if !utils.ExistsInSlice(name, allTeamsInFeathers) {
			return errors.Errorf("team %s does not exist in feathers", name)
		}
	}
	return nil
}

func (f *Feathers) GetAddressPoolByTeamNames(teamNames ...string) map[string][]string {
	wantedTeams := f.GetTeamsByNames(teamNames...)
	addressPool := make(map[string][]string, len(f.GetAllContactTypes()))
	for _, team := range wantedTeams {
		addressPool[team.ContactType] = append(addressPool[team.ContactType], team.Addresses...)
	}
	return addressPool
}

// validate checks that a team is set up correctly and contains all the required fields
func (t *Team) validate() error {
	// Check that none of the required fields are empty
	if t.Name == "" {
		return errors.New("no team name found")
	}
	if t.Addresses == nil {
		return errors.Errorf("no addresses for team %s", t.Name)
	}
	if t.ContactType == "" {
		return errors.Errorf("no contactType for team %s", t.Name)
	}
	if t.APIKey == "" {
		return errors.Errorf("no APIKey for team %s", t.Name)
	}

	// We should check that Peacock actually supports the contact type
	if exists := utils.ExistsInSlice(t.ContactType, models.Valid); !exists {
		return errors.Errorf("team %s has an invalid contact type of %s", t.Name, t.ContactType)
	}

	// We should check that the addresses conform to the contact type
	slackRegex, err := regexp.Compile(slackChannelIDRegex)
	if err != nil {
		return err
	}
	for _, address := range t.Addresses {
		switch t.ContactType {
		case models.Slack:
			match := slackRegex.MatchString(address)
			if !match {
				return errors.Errorf("failed to parse slack channel ID %s for team %s", address, t.Name)
			}
		}
	}
	return nil
}
