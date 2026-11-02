package feathers_test

import (
	"github.com/spring-financial-group/peacock/pkg/feathers"
	"github.com/spring-financial-group/peacock/pkg/models"
	"github.com/spring-financial-group/peacock/pkg/utils"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	"os"
	"path/filepath"
	"testing"
)

func Test_GetFeathersFromFile_Validate(t *testing.T) {
	testCases := []struct {
		name           string
		expectedConfig models.Feathers
		shouldError    bool
	}{
		{
			name: "Passing",
			expectedConfig: models.Feathers{
				Teams: []models.Team{
					{
						Name:        "infrastructure",
						ContactType: "slack",
						APIKey:      "9e7a455e-39f4-489b-b9ee-dd54d03c576e",
						Addresses:   []string{"C02BA9FHMD0"},
					},
					{
						Name:        "AllDevs",
						APIKey:      "24cacb0f-e186-4766-a444-14f028db63bd",
						ContactType: "slack",
						Addresses:   []string{"CLHLDNT9Q"},
					},
				},
			},
			shouldError: false,
		},
		{
			name: "InvalidContactType",
			expectedConfig: models.Feathers{
				Teams: []models.Team{
					{
						Name:        "infrastructure",
						APIKey:      "9e7a455e-39f4-489b-b9ee-dd54d03c576e",
						ContactType: "MorseCode",
						Addresses:   []string{"..-..-.--"},
					},
				},
			},
			shouldError: true,
		},
		{
			name: "SlackIDTooLong",
			expectedConfig: models.Feathers{
				Teams: []models.Team{
					{
						Name:        "infrastructure",
						APIKey:      "9e7a455e-39f4-489b-b9ee-dd54d03c576e",
						ContactType: "slack",
						Addresses:   []string{"C02DA9QHMD023"},
					},
				},
			},
			shouldError: true,
		},
		{
			name: "LowerCaseInSlackID",
			expectedConfig: models.Feathers{
				Teams: []models.Team{
					{
						Name:        "infrastructure",
						APIKey:      "9e7a455e-39f4-489b-b9ee-dd54d03c576e",
						ContactType: "slack",
						Addresses:   []string{"C02BA9QhMD"},
					},
				},
			},
			shouldError: true,
		},
		{
			name: "SlackIDTooShort",
			expectedConfig: models.Feathers{
				Teams: []models.Team{
					{
						Name:        "infrastructure",
						APIKey:      "9e7a455e-39f4-489b-b9ee-dd54d03c576e",
						ContactType: "slack",
						Addresses:   []string{"C02BA9QH"},
					},
				},
			},
			shouldError: true,
		},
		{
			name: "SlackIDWithNonAlphanumerics",
			expectedConfig: models.Feathers{
				Teams: []models.Team{
					{
						Name:        "infrastructure",
						APIKey:      "9e7a455e-39f4-489b-b9ee-dd54d03c576e",
						ContactType: "slack",
						Addresses:   []string{"C02!?/QHM"},
					},
				},
			},
			shouldError: true,
		},
		{
			name: "NoTeamName",
			expectedConfig: models.Feathers{
				Teams: []models.Team{
					{
						APIKey:      "9e7a455e-39f4-489b-b9ee-dd54d03c576e",
						ContactType: "email",
						Addresses:   []string{"sam.morse-dash.com"},
					},
				},
			},
			shouldError: true,
		},
		{
			name: "NoContactType",
			expectedConfig: models.Feathers{
				Teams: []models.Team{
					{
						Name:      "infrastructure",
						APIKey:    "9e7a455e-39f4-489b-b9ee-dd54d03c576e",
						Addresses: []string{"sam.morse-dash.com"},
					},
				},
			},
			shouldError: true,
		},
		{
			name: "MultipleTeams",
			expectedConfig: models.Feathers{
				Teams: []models.Team{
					{
						Name:        "infrastructure",
						ContactType: "slack",
						APIKey:      "9e7a455e-39f4-489b-b9ee-dd54d03c576e",
						Addresses:   []string{"C02BA9QHMD0"},
					},
					{
						Name:        "machine-learning",
						APIKey:      "eb7c0ee7-4ec2-474c-855f-51ab9c181cfa",
						ContactType: "slack",
						Addresses:   []string{"C02BA9QHMD0"},
					},
				},
			},
			shouldError: false,
		},
		{
			name: "DuplicateTeamNames",
			expectedConfig: models.Feathers{
				Teams: []models.Team{
					{
						Name:        "infrastructure",
						ContactType: "slack",
						APIKey:      "9e7a455e-39f4-489b-b9ee-dd54d03c576e",
						Addresses:   []string{"C02BA9QHMD0"},
					},
					{
						Name:        "infrastructure",
						APIKey:      "eb7c0ee7-4ec2-474c-855f-51ab9c181cfa",
						ContactType: "slack",
						Addresses:   []string{"C02BA9QHMD0"},
					},
				},
			},
			shouldError: true,
		},
		{
			name: "DuplicateAPIKeys",
			expectedConfig: models.Feathers{
				Teams: []models.Team{
					{
						Name:        "infrastructure",
						ContactType: "slack",
						APIKey:      "9e7a455e-39f4-489b-b9ee-dd54d03c576e",
						Addresses:   []string{"C02BA9QHMD0"},
					},
					{
						Name:        "infrastructure",
						APIKey:      "9e7a455e-39f4-489b-b9ee-dd54d03c576e",
						ContactType: "slack",
						Addresses:   []string{"C02BA9QHMD0"},
					},
				},
			},
			shouldError: true,
		},
		{
			name: "TeamWithNoneContactType",
			expectedConfig: models.Feathers{
				Teams: []models.Team{
					{
						Name:        "infrastructure",
						ContactType: "none",
						APIKey:      "9e7a455e-39f4-489b-b9ee-dd54d03c576e",
						Addresses:   []string{},
					},
				},
			},
			shouldError: false,
		},
		{
			name: "TeamWithNoneContactTypeButAddresses",
			expectedConfig: models.Feathers{
				Teams: []models.Team{
					{
						Name:        "infrastructure",
						ContactType: "none",
						APIKey:      "9e7a455e-39f4-489b-b9ee-dd54d03c576e",
						Addresses: []string{
							"SomeAddressThatShouldn'tExist",
						},
					},
				},
			},
			shouldError: true,
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
		t.Run(tt.name, func(t *testing.T) {
			uc := feathers.NewUseCase()

			bytes, err := yaml.Marshal(tt.expectedConfig)
			if err != nil {
				panic(err)
			}
			err = os.WriteFile(testPath, bytes, 0775)
			if err != nil {
				panic(err)
			}

			actualConfig, err := uc.GetFeathersFromFile()
			if tt.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectedConfig, *actualConfig)

			err = os.RemoveAll(testPath)
			if err != nil {
				panic(err)
			}
		})
	}

	err = os.RemoveAll(baseDir)
	if err != nil {
		panic(err)
	}
}

func Test_ValidateFeathers_Diffs(t *testing.T) {
	validTeams := models.Teams{
		{
			Name:        "infrastructure",
			ContactType: "slack",
			APIKey:      "9e7a455e-39f4-489b-b9ee-dd54d03c576e",
			Addresses:   []string{"C02BA9QHMD0"},
		},
	}

	testCases := []struct {
		name        string
		diffs       models.DiffConfigs
		shouldError bool
	}{
		{
			name: "ValidDiffs",
			diffs: models.DiffConfigs{
				{
					Paths:  []string{"config-root/**/*-schema-cm.yaml"},
					Format: models.DiffFormatConfigMap,
				},
			},
			shouldError: false,
		},
		{
			name:        "NoPaths",
			diffs:       models.DiffConfigs{{Format: models.DiffFormatText}},
			shouldError: true,
		},
		{
			name:        "EmptyPaths",
			diffs:       models.DiffConfigs{{Paths: []string{}, Format: models.DiffFormatText}},
			shouldError: true,
		},
		{
			name:        "InvalidFormat",
			diffs:       models.DiffConfigs{{Paths: []string{"*.yaml"}, Format: models.DiffFormat("yaml")}},
			shouldError: true,
		},
		{
			name:        "InvalidGlob",
			diffs:       models.DiffConfigs{{Paths: []string{"config-root/**/[.yaml"}, Format: models.DiffFormatText}},
			shouldError: true,
		},
		{
			name:        "ValidTeam",
			diffs:       models.DiffConfigs{{Paths: []string{"*.yaml"}, Format: models.DiffFormatText, Team: "infrastructure"}},
			shouldError: false,
		},
		{
			name:        "UnknownTeam",
			diffs:       models.DiffConfigs{{Paths: []string{"*.yaml"}, Format: models.DiffFormatText, Team: "DoesNotExist"}},
			shouldError: true,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			f := &models.Feathers{Teams: validTeams, Diffs: tt.diffs}
			err := feathers.NewUseCase().ValidateFeathers(f)
			if tt.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
