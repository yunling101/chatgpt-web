package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	GPT3Dot5Turbo0301  = "gpt-3.5-turbo-0301"
	GPT3Dot5Turbo      = "gpt-3.5-turbo"
	GPT3TextDavinci003 = "text-davinci-003"
)

// Session is a session created to communicate with OpenAI.
type Session struct {
	OrganizationID string
	Method         string
	HTTPClient     *http.Client
	apiKey         string
}

func NewSession(apiKey, method string) *Session {
	return &Session{
		apiKey: apiKey,
		Method: method,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (s *Session) MakeRequest(ctx context.Context, endpoint string, input interface{}) (response *http.Response, err error) {
	var (
		req *http.Request
		er  error
	)
	if input == nil {
		req, er = http.NewRequestWithContext(ctx, s.Method, endpoint, nil)
	} else {
		buf, er := json.Marshal(input)
		if er != nil {
			err = fmt.Errorf("marshal input error: %s", er.Error())
			return
		}
		req, er = http.NewRequestWithContext(ctx, s.Method, endpoint, bytes.NewReader(buf)) // http.MethodPost
	}
	if er != nil {
		err = fmt.Errorf("newRequestWithContext error: %s", er.Error())
		return
	}
	req.Header.Set("Content-Type", "application/json")
	return s.sendRequest(req)
}

func (s *Session) NewStreamRequest(
	ctx context.Context, endpoint string, input interface{}) (response *http.Response, err error) {
	buf, err := json.Marshal(input)
	if err != nil {
		err = fmt.Errorf("marshal input error: %s", err.Error())
		return
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(buf))
	if err != nil {
		err = fmt.Errorf("newRequestWithContext error: %s", err.Error())
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")
	return s.sendRequest(req)
}

func (s *Session) sendRequest(req *http.Request) (*http.Response, error) {
	if s.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+s.apiKey)
	}
	if s.OrganizationID != "" {
		req.Header.Set("OpenAI-Organization", s.OrganizationID)
	}
	return s.HTTPClient.Do(req)
}
