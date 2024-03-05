package controller

import (
	"encoding/json"

	"github.com/aglide100/ai-test/pkg/logger"
	"github.com/aglide100/ai-test/pkg/model"
	"github.com/gorilla/websocket"
)

func (c *WsController) HandleReceiveMessage(clientID string, conn *websocket.Conn) {
	defer func() {
		c.DisconnectClient(clientID, conn)
	}()
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			return
		}
		
		var data map[string]interface{}
		if err := json.Unmarshal(msg, &data); err != nil {
			logger.Error(err.Error())
			continue
		}

		var req model.ReceiveMessage

		command, ok := data["command"].(string)
		if !ok { 
			logger.Error(err.Error())
		}
 
		payload, err := json.Marshal(data["payload"])
		if err != nil {
			logger.Error(err.Error())
		}

		req.Payload = payload
		req.Command = model.ReceiveMessageType(command)
		var resp *model.Message

		client, ok := c.clients.Get(clientID)
		if !ok {
			logger.Error("Can't get client")
		}

		switch req.Command {
		case model.ReceiveMessageTypeDoneJob:
			resp = c.handleDoneJob(client, req.Payload)
		case model.ReceiveMessageTypeError:
			resp = c.handleError(client, req.Payload)
		case model.ReceiveMessageStillRun:
			resp = c.handleStillRun(client, req.Payload)
		default:
			logger.Error("Invalid Type")
			resp = model.NewErrorMessage(model.ErrMsgInvalidType)
		}

		if resp == nil {
			continue
		}

		respMsg, err := json.Marshal(resp)
		if err != nil {
			continue
		}

		if err := c.sendMessage(conn, respMsg); err != nil {
			continue
		}
	}
}


