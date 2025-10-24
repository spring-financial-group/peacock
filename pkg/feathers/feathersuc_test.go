package feathers_test

import (
	"github.com/spring-financial-group/peacock/pkg/feathers"
	"github.com/spring-financial-group/peacock/pkg/models"
	"github.com/spring-financial-group/peacock/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
				require.Error(t, err)
			} else {
				require.NoError(t, err)
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
