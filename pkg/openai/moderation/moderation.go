package moderation

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/yunling101/chatgpt-web/pkg/openai"
)

const (
	defaultListEndpoint = "https://api.openai.com/v1/models"
)

type Client struct {
	s            *openai.Session
	ListEndpoint string
}

type Response struct {
	Object string   `json:"object,omitempty"`
	Data   []object `json:"data,omitempty"`
}

type object struct {
	ID      string `json:"id,omitempty"`
	Object  string `json:"object,omitempty"`
	OwnedBy string `json:"owned_by,omitempty"`
	Root    string `json:"root,omitempty"`
}

func NewClient(session *openai.Session) *Client {
	return &Client{
		s:            session,
		ListEndpoint: defaultListEndpoint,
	}
}

func (c *Client) List(ctx context.Context) (response Response, err error) {
	req, err := c.s.MakeRequest(ctx, c.ListEndpoint, nil)
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
	err = json.Unmarshal(reader, &response)
	return
}
