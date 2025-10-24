package webhook

import (
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/spring-financial-group/peacock/pkg/markdown"
	httputils "github.com/spring-financial-group/peacock/pkg/utils/http_utils"
)

type Client struct {
	url    string
	token  string
	secret string
}

func NewClient(url, authToken, secret string) *Client {
	return &Client{
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

func (h *Client) Send(content, subject string, addresses []string) error {
	postReq := postRequest{
		Body:      markdown.ConvertToHTML(content),
		Subject:   subject,
		Addresses: addresses,
	}
	data, err := json.Marshal(postReq)
	if err != nil {
		return err
	}

	req, err := httputils.GenerateAuthenticatedPostRequest(h.url, h.token, httputils.SignMessage(data, h.secret), data)
	if err != nil {
		return err
	}
	_, err = httputils.DoRequestAndCatchUnsuccessful(req)
	if err != nil {
		return errors.Wrap(err, "failed to post messages")
	}
	return nil
}
