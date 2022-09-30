package http_utils

import (
	"bytes"
	"fmt"
	"github.com/pkg/errors"
	"net/http"
)

const (
	POST = "POST"

	ContentType     = "Content-Type"
	ApplicationJSON = "application/json"
	Authorization   = "Authorization"
)

func DoRequestAndCatchUnsuccessful(request *http.Request) (*http.Response, error) {
	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return resp, errors.New(fmt.Sprintf("Status code indicated failure %s", resp.Status))
	}
	return resp, nil
}

// GeneratePostRequest creates a POST http.Request. If a token is passed then it is added to the
// Authorization header.
func GeneratePostRequest(url, token string, body []byte) (*http.Request, error) {
	// create the request
	req, err := http.NewRequest(POST, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	// add content type
	req.Header.Add(ContentType, ApplicationJSON)
	if token != "" {
		req.Header.Add(Authorization, token)
	}
	return req, nil
}
