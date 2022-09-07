package email

import (
	"github.com/spring-financial-group/peacock/pkg/domain"
	gomail "gopkg.in/mail.v2"
)

const (
	to           = "to"
	subject      = "Subject"
	markdownMIME = "text/markdown"
)

type handler struct {
	dialer *gomail.Dialer
}

func NewEmailHandler(port int, host, username, password string) (domain.MessageHandler, error) {
	h := &handler{
		dialer: gomail.NewDialer(host, port, username, password),
	}
	return h, nil
}

func (h *handler) Send(content string, addresses []string) error {
	msg := gomail.NewMessage()
	msg.SetHeader(subject, "Peacock release notes")
	msg.SetBody(markdownMIME, content)
	for _, a := range addresses {
		msg.SetAddressHeader(to, a, "")
	}
	return h.dialer.DialAndSend(msg)
}
