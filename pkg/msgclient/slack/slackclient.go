package slack

import (
	"github.com/slack-go/slack"
	"github.com/spring-financial-group/peacock/pkg/markdown"
)

type Client struct {
	slack *slack.Client
}

func NewClient(token string) *Client {
	return &Client{
		slack: slack.New(token),
	}
}

func (c *Client) Send(content, _ string, addresses []string) error {
	content = markdown.ConvertToSlack(content)
	for _, address := range addresses {
		_, _, err := c.slack.PostMessage(
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
