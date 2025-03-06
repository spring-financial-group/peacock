package msgclients

import (
	"github.com/spring-financial-group/peacock/pkg/domain"
	"github.com/spring-financial-group/peacock/pkg/domain/mocks"
	"github.com/spring-financial-group/peacock/pkg/models"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	infraTeam = models.Team{
		Name:        "infrastructure",
		ContactType: models.Slack,
		Addresses:   []string{"#SlackAdd1", "#SlackAdd2"},
	}
	devsTeam = models.Team{
		Name:        "devs",
		ContactType: models.Slack,
		Addresses:   []string{"#SlackAdd3", "#SlackAdd4"},
	}
	supportTeam = models.Team{
		Name:        "support",
		ContactType: models.Webhook,
		Addresses:   []string{"Webhook1", "Webhook2"},
	}
	productTeam = models.Team{
		Name:        "product",
		ContactType: models.Webhook,
		Addresses:   []string{"Webhook3", "Webhook4"},
	}
	testingTeam = models.Team{
		Name:        "testing",
		ContactType: models.None,
		Addresses:   nil,
	}
	allTeams = models.Teams{
		infraTeam,
		devsTeam,
		supportTeam,
		productTeam,
		testingTeam,
	}
)

func TestHandler_SendMessage(t *testing.T) {
	slack := mocks.NewMessageClient(t)
	webhook := mocks.NewMessageClient(t)

	handler := &Handler{Clients: map[string]domain.MessageClient{
		models.Slack:   slack,
		models.Webhook: webhook,
	}}

	testCases := []struct {
		name         string
		inputMessage models.ReleaseNote
	}{
		{
			name: "Default",
			inputMessage: models.ReleaseNote{
				Teams:   allTeams,
				Content: "Test message content",
			},
		},
		{
			name: "NoneTeam",
			inputMessage: models.ReleaseNote{
				Teams:   allTeams,
				Content: "Test message content",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			slack.On("Send", tc.inputMessage.Content, "", []string{"#SlackAdd1", "#SlackAdd2", "#SlackAdd3", "#SlackAdd4"}).Return(nil)
			webhook.On("Send", tc.inputMessage.Content, "", []string{"Webhook1", "Webhook2", "Webhook3", "Webhook4"}).Return(nil)

			err := handler.SendReleaseNotes("", []models.ReleaseNote{tc.inputMessage})
			assert.NoError(t, err)
		})
	}
}
