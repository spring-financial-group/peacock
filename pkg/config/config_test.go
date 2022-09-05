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

func TestGetTeamByName(t *testing.T) {
	testCases := []struct {
		name          string
		inputTeamName string
		teams         []config.Team
		expectedTeam  *config.Team
	}{
		{
			name:          "Passing",
			inputTeamName: "infrastructure",
			teams: []config.Team{
				{Name: "infrastructure"},
			},
			expectedTeam: &config.Team{Name: "infrastructure"},
		},
		{
			name:          "MultipleTeams",
			inputTeamName: "infrastructure",
			teams: []config.Team{
				{Name: "infrastructure"},
				{Name: "ml"},
				{Name: "allDevs"},
			},
			expectedTeam: &config.Team{Name: "infrastructure"},
		},
		{
			name:          "NoTeamByThatName",
			inputTeamName: "NoTeam",
			teams: []config.Team{
				{Name: "infrastructure"},
				{Name: "ml"},
				{Name: "allDevs"},
			},
			expectedTeam: nil,
		},
		{
			name:          "NoTeams",
			inputTeamName: "infrastructure",
			teams:         []config.Team{},
			expectedTeam:  nil,
		},
	}

	for _, tt := range testCases {
		cfg := config.Config{Teams: tt.teams}

		t.Run(tt.name, func(t *testing.T) {
			actualTeam := cfg.GetTeamByName(tt.inputTeamName)
			assert.Equal(t, tt.expectedTeam, actualTeam)
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
