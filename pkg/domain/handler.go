package domain

type MessageHandler interface {
	// Send sends a message (content) to multiple addresses
	Send(content string, addresses []string) error
}
