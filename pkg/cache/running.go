package cache

import (
	"sync"
	"time"
)

type RunnerCache struct {
	clients *Cache
}

func NewRunnerCache(duration time.Duration, mutex *sync.Mutex) *RunnerCache {
	return &RunnerCache{
		clients: NewCache(duration, mutex, false),
	}
}

func (c *RunnerCache) Len() int {
	return c.clients.Len()
}

func (c *RunnerCache) Set(clientUUID string, jobId string) {
	c.clients.Set(clientUUID, jobId)
}

func (c *RunnerCache) Get(clientUUID string) (string, bool) {
	res, found := c.clients.Get(clientUUID)
	if (!found) {
		return "", false
	}

	jobId := res.(string)

	return jobId, true
}
