package feathers_test

import (
	"github.com/spring-financial-group/peacock/pkg/feathers"
	"github.com/spring-financial-group/peacock/pkg/utils"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	testCases := []struct {
		name           string
		expectedConfig feathers.Feathers
		shouldError    bool
	}{
		{
			name: "Passing",
			expectedConfig: feathers.Feathers{
				Teams: []feathers.Team{
					{
						Name:        "infrastructure",
						ContactType: "slack",
						Addresses:   []string{"C02BA9FHMD0"},
					},
					{
						Name:        "AllDevs",
						ContactType: "slack",
						Addresses:   []string{"CLHLDNT9Q"},
					},
				},
			},
			shouldError: false,
		},
		{
			name: "InvalidContactType",
			expectedConfig: feathers.Feathers{
				Teams: []feathers.Team{
					{
						Name:        "infrastructure",
						ContactType: "MorseCode",
						Addresses:   []string{"..-..-.--"},
					},
				},
			},
			shouldError: true,
		},
		{
			name: "ValidEmailAddress",
			expectedConfig: feathers.Feathers{
				Teams: []feathers.Team{
					{
						Name:        "infrastructure",
						ContactType: "email",
						Addresses:   []string{"sam.morse@dash.com"},
					},
				},
			},
			shouldError: false,
		},
		{
			name: "InvalidEmailAddress",
			expectedConfig: feathers.Feathers{
				Teams: []feathers.Team{
					{
						Name:        "infrastructure",
						ContactType: "email",
						Addresses:   []string{"sam.morse-dash.com"},
					},
				},
			},
			shouldError: true,
		},
		{
			name: "SlackIDTooLong",
			expectedConfig: feathers.Feathers{
				Teams: []feathers.Team{
					{
						Name:        "infrastructure",
						ContactType: "slack",
						Addresses:   []string{"C02DA9QHMD023"},
					},
				},
			},
			shouldError: true,
		},
		{
			name: "LowerCaseInSlackID",
			expectedConfig: feathers.Feathers{
				Teams: []feathers.Team{
					{
						Name:        "infrastructure",
						ContactType: "slack",
						Addresses:   []string{"C02BA9QhMD"},
					},
				},
			},
			shouldError: true,
		},
		{
			name: "SlackIDTooShort",
			expectedConfig: feathers.Feathers{
				Teams: []feathers.Team{
					{
						Name:        "infrastructure",
						ContactType: "slack",
						Addresses:   []string{"C02BA9QH"},
					},
				},
			},
			shouldError: true,
		},
		{
			name: "SlackIDWithNonAlphanumerics",
			expectedConfig: feathers.Feathers{
				Teams: []feathers.Team{
					{
						Name:        "infrastructure",
						ContactType: "slack",
						Addresses:   []string{"C02!?/QHM"},
					},
				},
			},
			shouldError: true,
		},
		{
			name: "NoTeamName",
			expectedConfig: feathers.Feathers{
				Teams: []feathers.Team{
					{
						ContactType: "email",
						Addresses:   []string{"sam.morse-dash.com"},
					},
				},
			},
			shouldError: true,
		},
		{
			name: "NoContactType",
			expectedConfig: feathers.Feathers{
				Teams: []feathers.Team{
					{
						Name:      "infrastructure",
						Addresses: []string{"sam.morse-dash.com"},
					},
				},
			},
			shouldError: true,
		},
		{
			name: "MultipleTeams",
			expectedConfig: feathers.Feathers{
				Teams: []feathers.Team{
					{
						Name:        "infrastructure",
						ContactType: "slack",
						Addresses:   []string{"C02BA9QHMD0"},
					},
					{
						Name:        "machine-learning",
						ContactType: "slack",
						Addresses:   []string{"C02BA9QHMD0"},
					},
				},
			},
			shouldError: false,
		},
	}

	baseDir, fullPath, err := utils.CreateTestDir(".peacock")
	if err != nil {
		panic(err)
	}
	testPath := filepath.Join(fullPath, "feathers.yaml")
	err = os.Chdir(baseDir)
	if err != nil {
		panic(err)
	}

	for _, tt := range testCases {
		bytes, err := yaml.Marshal(tt.expectedConfig)
		if err != nil {
			panic(err)
		}
		err = os.WriteFile(testPath, bytes, 0775)
		if err != nil {
			panic(err)
		}

		t.Run(tt.name, func(t *testing.T) {
			actualConfig, err := feathers.LoadConfig()
			if tt.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectedConfig, *actualConfig)
		})

		err = os.RemoveAll(testPath)
		if err != nil {
			panic(err)
		}
	}

	err = os.RemoveAll(baseDir)
	if err != nil {
		panic(err)
	}
}

func TestGetTeamsByNames(t *testing.T) {
	testCases := []struct {
		name           string
		inputTeamNames []string
		teams          []feathers.Team
		expectedTeams  []feathers.Team
	}{
		{
			name:           "Passing",
			inputTeamNames: []string{"infrastructure"},
			teams: []feathers.Team{
				{Name: "infrastructure"},
			},
			expectedTeams: []feathers.Team{{Name: "infrastructure"}},
		},
		{
			name:           "MultipleTeams",
			inputTeamNames: []string{"infrastructure"},
			teams: []feathers.Team{
				{Name: "infrastructure"},
				{Name: "ml"},
				{Name: "allDevs"},
			},
			expectedTeams: []feathers.Team{{Name: "infrastructure"}},
		},
		{
			name:           "NoTeamByThatName",
			inputTeamNames: []string{"DS"},
			teams: []feathers.Team{
				{Name: "infrastructure"},
				{Name: "ml"},
				{Name: "allDevs"},
			},
			expectedTeams: []feathers.Team(nil),
		},
		{
			name:           "NoTeams",
			inputTeamNames: []string{"infrastructure"},
			teams:          []feathers.Team{},
			expectedTeams:  []feathers.Team(nil),
		},
	}

	for _, tt := range testCases {
		cfg := feathers.Feathers{Teams: tt.teams}

		t.Run(tt.name, func(t *testing.T) {
			actualTeam := cfg.GetTeamsByNames(tt.inputTeamNames...)
			assert.Equal(t, tt.expectedTeams, actualTeam)
		})
	}
}

func TestGetAllTeamNames(t *testing.T) {
	testCases := []struct {
		name          string
		teams         []feathers.Team
		expectedNames []string
	}{
		{
			name: "Passing",
			teams: []feathers.Team{
				{Name: "infrastructure"},
			},
			expectedNames: []string{"infrastructure"},
		},
		{
			name: "MultipleTeams",
			teams: []feathers.Team{
				{Name: "infrastructure"},
				{Name: "ml"},
				{Name: "allDevs"},
			},
			expectedNames: []string{"infrastructure", "ml", "allDevs"},
		},
	}

	for _, tt := range testCases {
		cfg := feathers.Feathers{Teams: tt.teams}

		t.Run(tt.name, func(t *testing.T) {
			actualNames := cfg.GetAllTeamNames()
			assert.Equal(t, tt.expectedNames, actualNames)
		})
	}
}
