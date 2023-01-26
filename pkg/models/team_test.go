package models_test

import (
	"github.com/spring-financial-group/peacock/pkg/models"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetTeamsByNames(t *testing.T) {
	testCases := []struct {
		name           string
		inputTeamNames []string
		teams          models.Teams
		expectedTeams  models.Teams
	}{
		{
			name:           "Passing",
			inputTeamNames: []string{"infrastructure"},
			teams: models.Teams{
				{Name: "infrastructure"},
			},
			expectedTeams: models.Teams{{Name: "infrastructure"}},
		},
		{
			name:           "MultipleTeams",
			inputTeamNames: []string{"infrastructure"},
			teams: models.Teams{
				{Name: "infrastructure"},
				{Name: "ml"},
				{Name: "allDevs"},
			},
			expectedTeams: models.Teams{{Name: "infrastructure"}},
		},
		{
			name:           "NoTeamByThatName",
			inputTeamNames: []string{"DS"},
			teams: models.Teams{
				{Name: "infrastructure"},
				{Name: "ml"},
				{Name: "allDevs"},
			},
			expectedTeams: models.Teams(nil),
		},
		{
			name:           "NoTeams",
			inputTeamNames: []string{"infrastructure"},
			teams:          models.Teams{},
			expectedTeams:  models.Teams(nil),
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			actualTeam := tt.teams.GetTeamsByNames(tt.inputTeamNames...)
			assert.Equal(t, tt.expectedTeams, actualTeam)
		})
	}
}

func TestGetAllTeamNames(t *testing.T) {
	testCases := []struct {
		name          string
		teams         models.Teams
		expectedNames []string
	}{
		{
			name: "Passing",
			teams: models.Teams{
				{Name: "infrastructure"},
			},
			expectedNames: []string{"infrastructure"},
		},
		{
			name: "MultipleTeams",
			teams: models.Teams{
				{Name: "infrastructure"},
				{Name: "ml"},
				{Name: "allDevs"},
			},
			expectedNames: []string{"infrastructure", "ml", "allDevs"},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			actualNames := tt.teams.GetAllTeamNames()
			assert.Equal(t, tt.expectedNames, actualNames)
		})
	}
}
