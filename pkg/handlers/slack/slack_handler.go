package slack

import (
	"github.com/slack-go/slack"
	"github.com/spring-financial-group/peacock/pkg/domain"
)

type handler struct {
	slack *slack.Client
}

func NewSlackHandler(token string) (domain.MessageHandler, error) {
	h := &handler{
		slack: slack.New(token),
	}
	return h, nil
}

func (h *handler) Send(content string, addresses []string) error {
	for _, address := range addresses {
		_, _, err := h.slack.PostMessage(
			address,
			slack.MsgOptionText(content, false),
			slack.MsgOptionAsUser(true),
		)
		if err != nil {
			return err
		}
	}
	return nil
}
