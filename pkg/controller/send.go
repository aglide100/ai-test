package controller

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/aglide100/ai-test/pkg/logger"
	"github.com/aglide100/ai-test/pkg/model"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

func MessageToJson(msg *model.Message) ([]byte, error) {
	obj, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}

	return obj, nil 
}

func (c *WsController) HandleSendMessage(ctx context.Context, client *model.Client, conn *websocket.Conn) {
	t := time.NewTicker(60 * time.Second)
	defer func() {
		t.Stop()
		c.DisconnectClient(client.UUID, conn)
	}()
	for {
		select {
		case msg := <-client.Send:
			msgToJson, err := MessageToJson(msg)
			if err != nil {
				log.Println(err)
				logger.Error("can't make msg to json", zap.Error(err))
				return
			}

			if err := c.sendMessage(conn, msgToJson); err != nil {
				logger.Error("can't send", zap.Error(err))
				return
			}
		case <-t.C:
			msg := &model.Message{
				MessageType: model.MessageTypeStillRun,
			}
			
			msgToJson, err := MessageToJson(msg)
			if err != nil {
				log.Println(err)
				logger.Error("can't make msg to json", zap.Error(err))
				return
			}

			if err := c.sendMessage(conn, msgToJson); err != nil {
				logger.Info("disconnect: ", zap.Any("uuid", client.UUID))
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

func (c *WsController) sendMessage(conn *websocket.Conn, msg []byte) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	w, err := conn.NextWriter(websocket.TextMessage)
	if err != nil {
		return err
	}
	w.Write(msg)
	if err := w.Close(); err != nil {
		return err
	}
	return nil
}
