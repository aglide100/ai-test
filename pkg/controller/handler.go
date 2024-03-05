package controller

import (
	"github.com/aglide100/ai-test/pkg/logger"
	"github.com/aglide100/ai-test/pkg/model"
	"go.uber.org/zap"
)

func (c *WsController) handleDoneJob(client *model.Client, msg []byte) *model.Message {
	jobId, _ := c.clients.GetCurrentJob(client.UUID)
	res, found := c.waitingChannels.Get(jobId)
	
	c.doneJob<-client.UUID
	
	c.taskAllocator.RemoveByJobIdInRunning(jobId)
	

	if !found {
		c.results.Set(jobId, msg)
	} else {
		responseChan := res.(chan []byte)
		responseChan <- msg
	}

	return nil
}

func (c *WsController) handleError(client *model.Client, msg []byte) *model.Message {
	logger.Info("Can't process job", zap.Any("msg", string(msg[:])))
	jobId, _ := c.clients.GetCurrentJob(client.UUID)
	res, found := c.waitingChannels.Get(jobId)
	
	c.doneJob<-client.UUID
	
	c.taskAllocator.RemoveByJobIdInRunning(jobId)

	if !found {
		c.results.Set(jobId, msg)
	} else {
		responseChan := res.(chan []byte)
		responseChan <- msg
	}

	return nil
}

func (c *WsController) handleStillRun(client *model.Client, msg []byte) *model.Message {
	if (string(msg[:]) == "false") {
		if (c.clients.IsSelect(client.UUID)) {
			logger.Info("client doesn't work, reallocate...", zap.Any("client uuid", client.UUID))
			
			c.taskAllocator.RemoveRunningByClientUUID(client.UUID)
			c.clients.Deselect(client.UUID)
		}
	} 

	return nil

}
