package domain

import (
	feather "github.com/spring-financial-group/peacock/pkg/feathers"
	"github.com/spring-financial-group/peacock/pkg/models"
)

type MessageHandler interface {
	SendMessages(feathers *feather.Feathers, messages []models.ReleaseNote) error
	IsInitialised(contactType string) bool
}

type MessageClient interface {
	// Send sends a message to multiple addresses with a subject
	Send(content, subject string, addresses []string) error
}
