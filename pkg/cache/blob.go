package cache

import (
	"sync"
	"time"

	"github.com/aglide100/ai-test/pkg/logger"
	"go.uber.org/zap"
)

type BlobCache struct {
	blobs *Cache
	isWait *Cache
}

func NewBlobCache(duration time.Duration, mutex *sync.Mutex) *BlobCache {
	blobCache := &BlobCache{
		blobs: NewCache(duration, mutex, false),
		isWait: NewCache(duration, mutex, false),
	}

	go blobCache.CleanUp()
	return  blobCache
}

func (c *BlobCache) Set(key string, data []byte, isWaitUntilGet bool) {
	c.blobs.Set(key, data)
	c.isWait.Set(key, isWaitUntilGet)
}

func (c *BlobCache) Get(key string) ([]byte, bool) {
	res, found := c.blobs.Get(key)
	if !found {
		return  nil, false
	}

	data := res.([]byte)

	// res, found = c.isWait.Get(key)
	// if !found {
	// 	logger.Info("something is wrong, please check detail...")
	// 	return data, true
	// }

	// isWait := res.(bool)
	// if (isWait) {
	// 	c.isWait.Set(key, false)
	// 	logger.Info("It's first time", zap.Any("key", key))
	// } else {
	// 	// c.Delete(key)
	// }

	return data, true
}

func (c *BlobCache) Delete(key string) {
	c.blobs.Delete(key)
	c.isWait.Delete(key)
}

func (c *BlobCache) CleanUp() {
	for {
		select {
		case <-time.After(c.blobs.duration):
			for key, item := range c.blobs.data {
				res, found := c.isWait.Get(key)

				if !found {
					logger.Info("Can't find blobs", zap.Any("key", key))
					c.Delete(key)
				} else {
					if time.Since(item.CreatedAt) > c.blobs.duration {
						isWait := res.(bool)
						if !isWait {
							c.Delete(key)
						} else {
							c.blobs.Get(key)
						}
						
					}
				}
				
			}
		}
	}
}