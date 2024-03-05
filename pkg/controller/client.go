package controller

import (
	"net/url"
	"strings"

	"github.com/aglide100/ai-test/pkg/logger"
	"github.com/aglide100/ai-test/pkg/model"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

func (c *WsController) DisconnectClient(clientID string, conn *websocket.Conn) {
	logger.Info("disconnect", zap.String("uuid", clientID))
	c.clients.Delete(clientID)
	
	c.taskAllocator.RemoveRunningByClientUUID(clientID)
	logger.Info("Current state", zap.Any("cli", c.clients.Len()), zap.Any("waitingChannels", c.waitingChannels.Len()))
	err := conn.Close()
	if err != nil {
		logger.Error("Can't close client", zap.Any("err", err))
	}
}

func (c *WsController) CreateNewClient(clientIP string, queryParams url.Values) *model.Client {
	client := model.NewClient(clientIP)
	modes, ok := queryParams["modes"]
	if !ok {
		modes = []string{}
	} else {
		modes = strings.Split(modes[0], ",")
	}

	client.Tags = modes
	

	c.clients.Set(client.UUID, client)
	logger.Info("New client", zap.String("uuid", client.UUID), zap.String("ip", clientIP), zap.Any("modes", client.Tags))
	
	return client
}