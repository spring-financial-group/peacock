package http_utils

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
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

// SignMessage uses HMAC & SHA256 hashing to sign a message
func SignMessage(msg []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(msg)
	return hex.EncodeToString(mac.Sum(nil))
}

// DoRequestAndCatchUnsuccessful sends a http request. If the response code != 200 then it returns an error.
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

// GeneratePostRequest creates a POST http.Request adding the token to the Authorization header
func GeneratePostRequest(url, token string, body []byte) (*http.Request, error) {
	// create the request
	req, err := http.NewRequest(POST, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	// add content type
	req.Header.Add(ContentType, ApplicationJSON)
	req.Header.Add(Authorization, token)
	return req, nil
}
