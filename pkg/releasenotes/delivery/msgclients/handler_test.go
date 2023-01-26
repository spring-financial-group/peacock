package msgclients

import (
	"github.com/spring-financial-group/peacock/pkg/domain"
	"github.com/spring-financial-group/peacock/pkg/domain/mocks"
	"github.com/spring-financial-group/peacock/pkg/models"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHandler_SendMessage(t *testing.T) {
	slack := mocks.NewMessageClient(t)
	webhook := mocks.NewMessageClient(t)

	handler := &Handler{Clients: map[string]domain.MessageClient{
		models.Slack:   slack,
		models.Webhook: webhook,
	}}

	testCases := []struct {
		name          string
		inputMessage  models.ReleaseNote
		inputFeathers *models.Feathers
	}{
		{
			name: "Default",
			inputFeathers: &models.Feathers{
				Teams: []models.Team{
					{Name: "Infrastructure", ContactType: models.Slack, Addresses: []string{"#SlackAdd1", "#SlackAdd2"}},
					{Name: "AllDevs", ContactType: models.Slack, Addresses: []string{"#SlackAdd3", "#SlackAdd4"}},
					{Name: "Product", ContactType: models.Webhook, Addresses: []string{"Webhook1", "Webhook2"}},
					{Name: "Support", ContactType: models.Webhook, Addresses: []string{"Webhook3", "Webhook4"}},
				},
			},
			inputMessage: models.ReleaseNote{
				TeamNames: []string{"Infrastructure", "AllDevs", "Product", "Support"},
				Content:   "Test message content",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			slack.On("Send", tc.inputMessage.Content, "", []string{"#SlackAdd1", "#SlackAdd2", "#SlackAdd3", "#SlackAdd4"}).Return(nil)
			webhook.On("Send", tc.inputMessage.Content, "", []string{"Webhook1", "Webhook2", "Webhook3", "Webhook4"}).Return(nil)

			err := handler.SendReleaseNotes(tc.inputFeathers, []models.ReleaseNote{tc.inputMessage})
			assert.NoError(t, err)
		})
	}
}
