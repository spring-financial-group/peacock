package config_test

import (
	"github.com/spring-financial-group/peacock/pkg/config"
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
		expectedConfig config.Config
		shouldError    bool
	}{
		{
			name: "Passing",
			expectedConfig: config.Config{
				Teams: []config.Team{
					{
						Name:        "infrastructure",
						ContactType: "slack",
						Addresses:   []string{"C02BA9QHMD0"},
					},
				},
			},
			shouldError: false,
		},
		{
			name: "InvalidContactType",
			expectedConfig: config.Config{
				Teams: []config.Team{
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
			expectedConfig: config.Config{
				Teams: []config.Team{
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
			expectedConfig: config.Config{
				Teams: []config.Team{
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
			expectedConfig: config.Config{
				Teams: []config.Team{
					{
						Name:        "infrastructure",
						ContactType: "slack",
						Addresses:   []string{"C02BA9QHMD023"},
					},
				},
			},
			shouldError: true,
		},
		{
			name: "SlackIDTooShort",
			expectedConfig: config.Config{
				Teams: []config.Team{
					{
						Name:        "infrastructure",
						ContactType: "slack",
						Addresses:   []string{"C02BA9QHM"},
					},
				},
			},
			shouldError: true,
		},
		{
			name: "SlackIDWithNonAlphanumerics",
			expectedConfig: config.Config{
				Teams: []config.Team{
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
			expectedConfig: config.Config{
				Teams: []config.Team{
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
			expectedConfig: config.Config{
				Teams: []config.Team{
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
			expectedConfig: config.Config{
				Teams: []config.Team{
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
	testPath := filepath.Join(fullPath, "config.yaml")
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
			actualConfig, err := config.LoadConfig()
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
		teams          []config.Team
		expectedTeams  []config.Team
	}{
		{
			name:           "Passing",
			inputTeamNames: []string{"infrastructure"},
			teams: []config.Team{
				{Name: "infrastructure"},
			},
			expectedTeams: []config.Team{{Name: "infrastructure"}},
		},
		{
			name:           "MultipleTeams",
			inputTeamNames: []string{"infrastructure"},
			teams: []config.Team{
				{Name: "infrastructure"},
				{Name: "ml"},
				{Name: "allDevs"},
			},
			expectedTeams: []config.Team{{Name: "infrastructure"}},
		},
		{
			name:           "NoTeamByThatName",
			inputTeamNames: []string{"DS"},
			teams: []config.Team{
				{Name: "infrastructure"},
				{Name: "ml"},
				{Name: "allDevs"},
			},
			expectedTeams: []config.Team(nil),
		},
		{
			name:           "NoTeams",
			inputTeamNames: []string{"infrastructure"},
			teams:          []config.Team{},
			expectedTeams:  []config.Team(nil),
		},
	}

	for _, tt := range testCases {
		cfg := config.Config{Teams: tt.teams}

		t.Run(tt.name, func(t *testing.T) {
			actualTeam := cfg.GetTeamsByNames(tt.inputTeamNames...)
			assert.Equal(t, tt.expectedTeams, actualTeam)
		})
	}
}

func TestGetAllTeamNames(t *testing.T) {
	testCases := []struct {
		name          string
		teams         []config.Team
		expectedNames []string
	}{
		{
			name: "Passing",
			teams: []config.Team{
				{Name: "infrastructure"},
			},
			expectedNames: []string{"infrastructure"},
		},
		{
			name: "MultipleTeams",
			teams: []config.Team{
				{Name: "infrastructure"},
				{Name: "ml"},
				{Name: "allDevs"},
			},
			expectedNames: []string{"infrastructure", "ml", "allDevs"},
		},
	}

	for _, tt := range testCases {
		cfg := config.Config{Teams: tt.teams}

		t.Run(tt.name, func(t *testing.T) {
			actualNames := cfg.GetAllTeamNames()
			assert.Equal(t, tt.expectedNames, actualNames)
		})
	}
}
