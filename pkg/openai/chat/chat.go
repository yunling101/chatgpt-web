package chat

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/yunling101/chatgpt-web/pkg/openai"
)

const (
	MessageRoleSystem    = "system"
	MessageRoleUser      = "user"
	MessageRoleAssistant = "assistant"
	defaultApiEndpoint   = "https://api.openai.com/v1/chat/completions"
)

type Client struct {
	s                  *openai.Session
	model              string
	defaultApiEndpoint string
}

func NewClient(session *openai.Session, model string) *Client {
	if model == "" {
		model = openai.GPT3Dot5Turbo
	}
	return &Client{
		s:                  session,
		model:              model,
		defaultApiEndpoint: defaultApiEndpoint,
	}
}

type Params struct {
	Model            string     `json:"model,omitempty"`
	Messages         []*Message `json:"messages,omitempty"`
	Stop             []string   `json:"stop,omitempty"`
	Stream           bool       `json:"stream,omitempty"`
	N                int        `json:"n,omitempty"`
	TopP             float64    `json:"top_n,omitempty"`
	Temperature      float64    `json:"temperature,omitempty"`
	MaxTokens        int        `json:"max_tokens,omitempty"`
	PresencePenalty  float64    `json:"presence_penalty,omitempty"`
	FrequencyPenalty float64    `json:"frequency_penalty,omitempty"`
	User             string     `json:"user,omitempty"`
}

type Response struct {
	ID      string    `json:"id,omitempty"`
	Object  string    `json:"object,omitempty"`
	Created int64     `json:"created,omitempty"`
	Choices []*Choice `json:"choices,omitempty"`
	Model   string    `json:"model,omitempty"`
}

type Choice struct {
	Delta        *Message `json:"delta,omitempty"`
	Index        int      `json:"index,omitempty"`
	FinishReason string   `json:"finish_reason,omitempty"`
	Message      *Message `json:"message,omitempty"`
}

type Message struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
}

func (c *Client) CreateCompletion(ctx context.Context, p *Params) (r Response, err error) {
	if p.Model == "" && c.model == "" {
		err = fmt.Errorf("%s", "model cannot be empty")
		return
	}
	if p.Model == "" {
		p.Model = c.model
	}

	req, err := c.s.MakeRequest(ctx, c.defaultApiEndpoint, p)
	if err != nil {
		err = fmt.Errorf("%s", err.Error())
		return
	}
	defer req.Body.Close()

	reader, er := ioutil.ReadAll(req.Body)
	if er != nil {
		err = fmt.Errorf("%s", er.Error())
		return
	}
	err = json.Unmarshal(reader, &r)
	return
}

func (c *Client) Recv(r []byte) (response Response, err error) {
	var headerData = []byte("data: ")
	n := bytes.TrimSpace(r)
	if !bytes.HasPrefix(n, headerData) {
		err = fmt.Errorf("%s", "stream has sent too many empty messages")
		return
	}
	n = bytes.TrimPrefix(n, headerData)
	if string(n) == "[DONE]" {
		err = io.EOF
		return
	}

	err = json.Unmarshal(n, &response)
	return
}

func (c *Client) CreateCompletionStream(ctx context.Context, p *Params) (reader *bufio.Reader, err error) {
	if p.Model == "" && c.model == "" {
		err = fmt.Errorf("%s", "model cannot be empty")
		return
	}
	if p.Model == "" {
		p.Model = c.model
	}
	req, err := c.s.NewStreamRequest(ctx, c.defaultApiEndpoint, p)
	if err != nil {
		err = fmt.Errorf("%s", err.Error())
		return
	}
	reader = bufio.NewReader(req.Body)
	return
}
