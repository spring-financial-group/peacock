package domain

import "github.com/spring-financial-group/peacock/pkg/models"

type MessageHandler interface {
	SendReleaseNotes(subject string, notes []models.ReleaseNote) error
	IsInitialised(contactType string) bool
}

type MessageClient interface {
	// Send sends a message to multiple addresses with a subject
	Send(content, subject string, addresses []string) error
}
