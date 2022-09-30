package webhook

import (
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/spring-financial-group/peacock/pkg/domain"
	"github.com/spring-financial-group/peacock/pkg/utils/http_utils"
)

type handler struct {
	url   string
	token string
}

func NewWebHookHandler(url, token string) domain.MessageHandler {
	return &handler{
		url:   url,
		token: token,
	}
}

type postRequest struct {
	Body        string
	Subject     string
	ToAddresses []string
}

func (h *handler) Send(content, subject string, addresses []string) error {
	postReq := postRequest{
		Body:        content,
		Subject:     subject,
		ToAddresses: addresses,
	}
	data, err := json.Marshal(postReq)
	if err != nil {
		return err
	}
	req, err := http_utils.GeneratePostRequest(h.url, h.token, data)
	if err != nil {
		return err
	}
	_, err = http_utils.DoRequestAndCatchUnsuccessful(req)
	if err != nil {
		return errors.Wrap(err, "failed to post messages")
	}
	return nil
}
