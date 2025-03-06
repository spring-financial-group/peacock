package models

const (
	Slack   = "slack"
	Webhook = "webhook"
	None    = "none"
)

var Valid = []string{Slack, Webhook, None}
