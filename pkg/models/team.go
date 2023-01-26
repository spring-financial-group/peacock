package models

import (
	"github.com/pkg/errors"
	"github.com/spring-financial-group/peacock/pkg/utils"
)

type Team struct {
	Name        string   `yaml:"name"`
	APIKey      string   `yaml:"apiKey"`
	ContactType string   `yaml:"contactType"`
	Addresses   []string `yaml:"addresses"`
}

type Teams []Team

func (ts Teams) GetTeamsByNames(names ...string) Teams {
	var teams Teams
	for _, name := range names {
		for _, t := range ts {
			if t.Name == name {
				teams = append(teams, t)
			}
		}
	}
	return teams
}

func (ts Teams) GetAllTeamNames() []string {
	var names []string
	for _, t := range ts {
		names = append(names, t.Name)
	}
	return names
}

func (ts Teams) GetAllContactTypes() []string {
	var types []string
	for _, t := range ts {
		types = append(types, t.ContactType)
	}
	return types
}

func (ts Teams) GetContactTypesByTeamNames(names ...string) []string {
	var types []string
	for _, team := range ts.GetTeamsByNames(names...) {
		types = append(types, team.ContactType)
	}
	return types
}

func (ts Teams) Exists(teamNames ...string) error {
	allTeamsInFeathers := ts.GetAllTeamNames()
	for _, name := range teamNames {
		if !utils.ExistsInSlice(name, allTeamsInFeathers) {
			return errors.Errorf("team %s does not exist in feathers", name)
		}
	}
	return nil
}

func (ts Teams) GetAddressPoolByTeamNames(teamNames ...string) map[string][]string {
	wantedTeams := ts.GetTeamsByNames(teamNames...)
	addressPool := make(map[string][]string, len(ts.GetAllContactTypes()))
	for _, team := range wantedTeams {
		addressPool[team.ContactType] = append(addressPool[team.ContactType], team.Addresses...)
	}
	return addressPool
}
