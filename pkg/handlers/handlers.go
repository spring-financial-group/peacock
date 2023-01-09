package handlers

import (
	"github.com/spring-financial-group/peacock/pkg/domain"
	"github.com/spring-financial-group/peacock/pkg/handlers/slack"
	"github.com/spring-financial-group/peacock/pkg/handlers/webhook"
)

const (
	Slack   = "slack"
	Webhook = "webhook"
)

var Valid = []string{Slack, Webhook}

// InitMessageHandlers returns a map of message handlers base on what values are available
func InitMessageHandlers(slackToken, webhookURL, webhookToken, webhookSecret string) map[string]domain.MessageHandler {
	handlers := make(map[string]domain.MessageHandler)
	if slackToken != "" {
		handlers[Slack] = slack.NewSlackHandler(slackToken)
	}
	if webhookURL != "" && webhookToken != "" && webhookSecret != "" {
		handlers[Webhook] = webhook.NewWebHookHandler(webhookURL, webhookToken, webhookSecret)
	}
	return handlers
}
