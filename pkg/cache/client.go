package cache

import (
	"sync"
	"time"

	"github.com/aglide100/ai-test/pkg/logger"
	"github.com/aglide100/ai-test/pkg/model"
	"github.com/aglide100/ai-test/pkg/queue"
)

type ClientCache struct {
	clients  *Cache
	selectedClient *Cache
	availableClient    *queue.PriorityQueue
	currentJob  *Cache
}

type SelectedClient struct {
	ClientUUID string
}

func NewClientCache(duration time.Duration, mutex *sync.Mutex) *ClientCache {
	return &ClientCache{
		clients:  NewCache(duration, mutex, false),
		selectedClient: NewCache(duration, mutex, false),
		availableClient:    queue.NewPriorityQueue(mutex),
		currentJob:  NewCache(duration, mutex, true),
	}
}

func (c *ClientCache) Len() int {
	return c.clients.Len()
}

func (c *ClientCache) GetCurrentJob(clientUUID string) (string, bool) {
	res, found := c.currentJob.Get(clientUUID)
	if !found {
		logger.Info("Can't find client, please check more detail")
		return "", found
	}

	jobId := res.(string)

	return jobId, found
}

func (c *ClientCache) IsSelect(clientUUID string) bool {
	res, found := c.selectedClient.Get(clientUUID)
	if !found {
		logger.Info("Can't find client")
		return false
	}

	isSelect := res.(bool)

	return isSelect
}

func (c *ClientCache) GetNotSelectedClient(jobId string, mode string) (*model.Client, bool) {
	if c.clients.Len() == 0 || c.availableClient.Len() == 0 {
		return nil, false
	}

	tmp := []*SelectedClient{}

	for {
		val, found := c.availableClient.Pop()
		if !found {
			break
		}

		sel := val.Value.(*SelectedClient)

		res, found := c.clients.Get(sel.ClientUUID)
		if !found {
			logger.Info("can't find client")
			break
		}
		cli := res.(*model.Client)

		item, exists := c.selectedClient.Get(sel.ClientUUID)

		busy := item.(bool)
		if exists {
			if !busy && cli.Contains(mode) {
				c.selectedClient.Set(sel.ClientUUID, true)
				c.currentJob.Set(sel.ClientUUID, jobId)

				res, found := c.clients.Get(sel.ClientUUID)
				if !found {
					logger.Info("Can't find client, please check logic")
				}

				cli := res.(*model.Client)

				return cli, true
			}
		}
		tmp = append(tmp, sel)
	}

	for _, val := range tmp {
		c.availableClient.Push(&queue.QueueItem{
			Value: val,
		})
	}

	return nil, false
}

func (c *ClientCache) Deselect(key string) bool {
	_, found := c.selectedClient.Get(key)
	if !found {
		return false
	}

	c.availableClient.Push(&queue.QueueItem{
		Value: &SelectedClient{
			ClientUUID: key,
		},
	})

	c.selectedClient.Set(key, false)
	c.currentJob.Delete(key)
	return true
}

func (c *ClientCache) Set(key string, value interface{}) {
	c.clients.Set(key, value)
	c.selectedClient.Set(key, false)
	c.availableClient.Push(&queue.QueueItem{
		Value: &SelectedClient{
			ClientUUID: key,
		},
	})
}

func (c *ClientCache) Get(key string) (*model.Client, bool) {
	res, found := c.clients.Get(key)
	if !found {
		return nil, found
	}

	client := res.(*model.Client)

	return client, found
}

func (c *ClientCache) Delete(key string) {
	res, found := c.clients.Get(key)
	if found {
		client := res.(*model.Client)
		close(client.Send)
	}

	c.clients.Delete(key)
	c.selectedClient.Delete(key)
	c.currentJob.Delete(key)

	tmp := []*SelectedClient{}

	for {
		val, found := c.availableClient.Pop()
		if !found {
			break
		}

		sel := val.Value.(*SelectedClient)
		if key == sel.ClientUUID {
			break
		}

		tmp = append(tmp, sel)
	}

	for _, val := range tmp {
		c.availableClient.Push(&queue.QueueItem{
			Value: val,
		})
	}
}
