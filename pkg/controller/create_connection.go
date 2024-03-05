package controller

import (
	"context"
	"net/http"

	"github.com/aglide100/ai-test/pkg/logger"
	"github.com/aglide100/ai-test/pkg/model"
	"go.uber.org/zap"
)


type SendClientID struct {
	ClientID string `json:"client_id"`
}

func (c *WsController) CreateConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	queryParams := r.URL.Query()
	token := queryParams.Get("token")	

	if (c.token != token) {
		logger.Info("token is invalid", zap.Any("token", token))
		return
	}

	client := c.CreateNewClient(ReadUserIP(r), queryParams)

	ctx := context.Background()

	go c.HandleSendMessage(ctx, client, conn)
	go c.HandleReceiveMessage(client.UUID, conn)

	msgToJson, err := MessageToJson(&model.Message{
		MessageType: model.MessageTypeNotifyClientID,
		Payload: &SendClientID{ClientID: client.UUID},
	})

	if err != nil {
		logger.Error("err", zap.Any("e", err.Error()))
		return
	}

	go func() {
		requestData := &model.RequestData{}
		c.readableRequest <- requestData
	}()
	
	c.sendMessage(conn, msgToJson)
}

func ReadUserIP(r *http.Request) string {
    IPAddress := r.Header.Get("X-Real-Ip")
    if IPAddress == "" {
        IPAddress = r.Header.Get("X-Forwarded-For")
    }
    if IPAddress == "" {
        IPAddress = r.RemoteAddr
    }
    return IPAddress
}