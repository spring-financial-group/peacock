package domain

type MessageHandler interface {
	// Send sends a message to multiple addresses with a subject
	Send(content, subject string, addresses []string) error
}
