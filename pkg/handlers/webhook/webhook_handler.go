package webhook

import (
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/spring-financial-group/peacock/pkg/domain"
	"github.com/spring-financial-group/peacock/pkg/markdown"
	"github.com/spring-financial-group/peacock/pkg/utils/http_utils"
)

type handler struct {
	url    string
	token  string
	secret string
}

func NewWebHookHandler(url, authToken, secret string) domain.MessageHandler {
	return &handler{
		url:    url,
		token:  authToken,
		secret: secret,
	}
}

type postRequest struct {
	Body      string   `json:"body"`
	Subject   string   `json:"subject"`
	Addresses []string `json:"addresses"`
}

func (h *handler) Send(content, subject string, addresses []string) error {
	postReq := postRequest{
		Body:      markdown.ConvertToHTML(content),
		Subject:   subject,
		Addresses: addresses,
	}
	data, err := json.Marshal(postReq)
	if err != nil {
		return err
	}

	req, err := http_utils.GenerateAuthenticatedPostRequest(h.url, h.token, http_utils.SignMessage(data, h.secret), data)
	if err != nil {
		return err
	}
	_, err = http_utils.DoRequestAndCatchUnsuccessful(req)
	if err != nil {
		return errors.Wrap(err, "failed to post messages")
	}
	return nil
}
