package http_utils

import (
	"bytes"
	"context"
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

	AuthorizationHeader = "Authorization"
	SignatureHeader     = "X-Signature-256"
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

// GenerateAuthenticatedPostRequest creates a POST http.Request adding a token to the AuthorizationHeader header
// and hash to the SignatureHeader
func GenerateAuthenticatedPostRequest(url, authToken, hash string, body []byte) (*http.Request, error) {
	// create the request
	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, POST, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	// add content type
	req.Header.Add(ContentType, ApplicationJSON)
	req.Header.Add(AuthorizationHeader, authToken)
	req.Header.Add(SignatureHeader, hash)
	return req, nil
}
