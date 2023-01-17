package msgclient

import (
	"github.com/spring-financial-group/peacock/pkg/domain"
	"github.com/spring-financial-group/peacock/pkg/domain/mocks"
	"github.com/spring-financial-group/peacock/pkg/feathers"
	"github.com/spring-financial-group/peacock/pkg/message"
	"github.com/spring-financial-group/peacock/pkg/models"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHandler_SendMessage(t *testing.T) {
	slack := mocks.NewMessageClient(t)
	webhook := mocks.NewMessageClient(t)

	handler := &Handler{clients: map[string]domain.MessageClient{
		models.Slack:   slack,
		models.Webhook: webhook,
	}}

	testCases := []struct {
		name          string
		inputMessage  message.Message
		inputFeathers *feathers.Feathers
	}{
		{
			name: "Default",
			inputFeathers: &feathers.Feathers{
				Teams: []feathers.Team{
					{Name: "Infrastructure", ContactType: models.Slack, Addresses: []string{"#SlackAdd1", "#SlackAdd2"}},
					{Name: "AllDevs", ContactType: models.Slack, Addresses: []string{"#SlackAdd3", "#SlackAdd4"}},
					{Name: "Product", ContactType: models.Webhook, Addresses: []string{"Webhook1", "Webhook2"}},
					{Name: "Support", ContactType: models.Webhook, Addresses: []string{"Webhook3", "Webhook4"}},
				},
			},
			inputMessage: message.Message{
				TeamNames: []string{"Infrastructure", "AllDevs", "Product", "Support"},
				Content:   "Test message content",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			slack.On("Send", tc.inputMessage.Content, "", []string{"#SlackAdd1", "#SlackAdd2", "#SlackAdd3", "#SlackAdd4"}).Return(nil)
			webhook.On("Send", tc.inputMessage.Content, "", []string{"Webhook1", "Webhook2", "Webhook3", "Webhook4"}).Return(nil)

			err := handler.SendMessages(tc.inputFeathers, []message.Message{tc.inputMessage})
			assert.NoError(t, err)
		})
	}
}
