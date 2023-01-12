package slack

import (
	"github.com/slack-go/slack"
	"github.com/spring-financial-group/peacock/pkg/markdown"
)

type Handler struct {
	slack *slack.Client
}

func NewSlackHandler(token string) *Handler {
	return &Handler{
		slack: slack.New(token),
	}
}

func (h *Handler) Send(content, _ string, addresses []string) error {
	content = markdown.ConvertToSlack(content)
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
