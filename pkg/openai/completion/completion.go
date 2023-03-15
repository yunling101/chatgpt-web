package completion

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/yunling101/chatgpt-web/pkg/openai"
)

//https://platform.openai.com/docs/api-reference/completions/create

const defaultCreateEndpoint = "https://api.openai.com/v1/completions"

type Client struct {
	s              *openai.Session
	model          string
	CreateEndpoint string
}

func NewClient(session *openai.Session, model string) *Client {
	return &Client{
		s:              session,
		model:          model,
		CreateEndpoint: defaultCreateEndpoint,
	}
}

type Params struct {
	Model            string   `json:"model,omitempty"`
	Prompt           []string `json:"prompt,omitempty"`
	Stop             []string `json:"stop,omitempty"`
	Suffix           string   `json:"suffix,omitempty"`
	Stream           bool     `json:"stream,omitempty"`
	Echo             bool     `json:"echo,omitempty"`
	MaxTokens        int      `json:"max_tokens,omitempty"`
	N                int      `json:"n,omitempty"`
	TopP             float64  `json:"top_n,omitempty"`
	Temperature      float64  `json:"temperature,omitempty"`
	LogProbs         int      `json:"logprobs,omitempty"`
	PresencePenalty  float64  `json:"presence_penalty,omitempty"`
	FrequencyPenalty float64  `json:"frequency_penalty,omitempty"`
	BestOf           int      `json:"best_of,omitempty"`
	User             string   `json:"user,omitempty"`
}

type Response struct {
	ID      string    `json:"id,omitempty"`
	Object  string    `json:"object,omitempty"`
	Created int64     `json:"created,omitempty"`
	Choices []*Choice `json:"choices,omitempty"`
	Model   string    `json:"model,omitempty"`
}

type Choice struct {
	Text         string `json:"text,omitempty"`
	Index        int    `json:"index,omitempty"`
	LogProbs     int    `json:"logprobs,omitempty"`
	FinishReason string `json:"finish_reason,omitempty"`
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

	req, err := c.s.NewStreamRequest(ctx, c.CreateEndpoint, p)
	if err != nil {
		err = fmt.Errorf("%s", err.Error())
		return
	}
	reader = bufio.NewReader(req.Body)
	return
}
