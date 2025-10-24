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

type UseCase struct {
}

func NewUseCase() *UseCase {
	return &UseCase{}
}

func (uc *UseCase) GetFeathersFromFile() (*models.Feathers, error) {
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
	return uc.GetFeathersFromBytes(data)
}

func (uc *UseCase) GetFeathersFromBytes(data []byte) (*models.Feathers, error) {
	feathers := new(models.Feathers)
	err := yaml.Unmarshal(data, &feathers)
	if err != nil {
		return nil, err
	}
	return feathers, uc.ValidateFeathers(feathers)
}

func (uc *UseCase) ValidateFeathers(f *models.Feathers) error {
	if f.Teams == nil {
		return errors.New("no teams found in feathers")
	}

	const name, apiKey = "Name", "APIKey"
	unique := map[string]map[string]bool{
		name:   make(map[string]bool),
		apiKey: make(map[string]bool),
	}

	for _, team := range f.Teams {
		// Check that the individual teams are set up correctly
		err := uc.validateTeam(team)
		if err != nil {
			return err
		}

		// Check the team names and API keys are unique
		if _, exists := unique[name][team.Name]; exists {
			return fmt.Errorf("duplicate team name found: %s", team.Name)
		}
		if _, exists := unique[apiKey][team.APIKey]; exists {
			return fmt.Errorf("duplicate apiKey found: %s", team.APIKey)
		}
		unique[name][team.Name] = true
		unique[apiKey][team.APIKey] = true
	}
	return nil
}

// validate checks that a team is set up correctly and contains all the required fields
func (uc *UseCase) validateTeam(t models.Team) error {
	// Check that none of the required fields are empty
	if t.Name == "" {
		return errors.New("no team name found")
	}
	if t.ContactType == "" {
		return errors.Errorf("no contactType for team %s", t.Name)
	}
	if len(t.Addresses) == 0 && t.ContactType != models.None {
		return errors.Errorf("no addresses for team %s", t.Name)
	}
	if len(t.Addresses) > 0 && t.ContactType == models.None {
		return errors.Errorf("addresses found for team %s with contactType of none", t.Name)
	}

	if t.APIKey == "" {
		return errors.Errorf("no APIKey for team %s", t.Name)
	}

	// We should check that Peacock actually supports the contact type
	if exists := utils.ExistsInSlice(t.ContactType, models.Valid); !exists {
		return errors.Errorf("team %s has an invalid contact type of %s", t.Name, t.ContactType)
	}

	// We should check that the addresses conform to the contact type
	slackRegex := regexp.MustCompile(slackChannelIDRegex)
	for _, address := range t.Addresses {
		if t.ContactType == models.Slack {
			match := slackRegex.MatchString(address)
			if !match {
				return errors.Errorf("failed to parse slack channel ID %s for team %s", address, t.Name)
			}
		}
	}
	return nil
}
