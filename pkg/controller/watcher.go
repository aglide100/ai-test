package controller

import (
	"time"

	"github.com/aglide100/ai-test/pkg/logger"
	"github.com/aglide100/ai-test/pkg/model"
	"go.uber.org/zap"
)

func (c *WsController) Watcher() {
	for {
		select {
		case key := <-c.doneJob:
			c.Done(key)
		case requestData := <-c.readableRequest:
			jobId := requestData.Data
			responseChan := requestData.ResponseChan

			if (jobId != "" && responseChan != nil) {
				c.waitingChannels.Set(jobId, responseChan)
			}

			len_cli := c.clients.Len()
			len_running := c.taskAllocator.LenRunning()

			if (len_cli > len_running) {
				logger.Info("allocate", zap.Any("cli", len_cli), zap.Any("running", len_running))
				c.Allocate()
			}
		}
	}
}

func (c *WsController) Done(clientUUID string) {
	c.clients.Deselect(clientUUID)

	logger.Info("Done job", zap.Any("clientUUID", clientUUID), zap.Any("len_cli", c.clients.Len()))
	c.Allocate()
}

func (c *WsController) Allocate() {
	job, found := c.taskAllocator.PopInWait()
	if (found && job != nil){
		logger.Info("job", zap.Any("job", job))
		cli, found := c.clients.GetNotSelectedClient(job.ID, job.Mode)
		if found {
			c.taskAllocator.AddInRunning(job, cli.UUID)
			
			msg := &model.Message{
				MessageType: model.MessageTypeGiveJob,
				Payload: job,
			}
			go func() {
				logger.Info("send to client", zap.Any("msg", msg))
				cli.Send <- msg
			}()
			
		} else {
			logger.Info("There's no client... retrying...", zap.Any("waiting", c.taskAllocator.LenWaiting()))
			time.Sleep(3 * time.Second)
			c.taskAllocator.AddInWait(job)
		}
	}
}
