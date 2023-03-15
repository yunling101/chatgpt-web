package route

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/yunling101/chatgpt-web/config"
	"github.com/yunling101/chatgpt-web/pkg/openai"
	"github.com/yunling101/chatgpt-web/pkg/openai/chat"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

type requestMessage struct {
	Index    int    `json:"index"`
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

// HandleWebSocket 升级 HTTP 连接为 WebSocket 连接
func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer conn.Close()

	// 读取客户端发送的消息并回复
	session := openai.NewSession(config.Token, "POST")
	chatMessage := make([]*chat.Message, 0)
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("failed to read message: %v", err.Error())
			break
		}

		var req requestMessage
		if err = json.Unmarshal(message, &req); err != nil {
			log.Printf("json request message error: %s", err.Error())
			break
		}
		if req.Question == "" {
			continue
		}

		chatMessage = append(chatMessage, &chat.Message{Role: chat.MessageRoleUser, Content: req.Question})
		client := chat.NewClient(session, openai.GPT3Dot5Turbo)
		reader, err := client.CreateCompletionStream(context.Background(), &chat.Params{
			MaxTokens: config.MaxToken,
			Messages:  chatMessage,
			Stream:    true,
		})
		if err != nil {
			log.Printf("Failed to complete: %s", err.Error())
			break
		}

		var (
			count  = 0
			answer string
		)
		for {
			nr, er := reader.ReadBytes('\n')
			if er != nil {
				break
			}
			response, err := client.Recv(nr)
			if err != nil {
				if err == io.EOF {
					break
				}
				continue
			}
			for _, v := range response.Choices {
				if count == 0 {
					if v.Delta.Content == "" || v.Delta.Content == "\n\n" {
						continue
					}
				}
				count++
				answer = answer + v.Delta.Content
				b, _ := json.Marshal(requestMessage{Answer: v.Delta.Content, Index: req.Index, Question: req.Question})
				if err := conn.WriteMessage(messageType, b); err != nil {
					log.Printf("failed to write message: %v", err)
					break
				}
			}
		}
		chatMessage = append(chatMessage, &chat.Message{Role: chat.MessageRoleAssistant, Content: answer})
		_ = conn.WriteMessage(messageType, []byte(`{"answer": "Done"}`))
	}
}
