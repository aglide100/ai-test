package controller

import (
	"net/http"
	"sync"

	"github.com/aglide100/ai-test/pkg/cache"
	"github.com/aglide100/ai-test/pkg/model"
	"github.com/aglide100/ai-test/pkg/queue"
	"github.com/gorilla/websocket"
)

type WsController struct {
	clients *cache.ClientCache
	results *cache.Cache
	taskAllocator *queue.TaskAllocator
	token string
	doneJob chan string
	waitingChannels *cache.Cache
	readableRequest chan *model.RequestData
	mutex    *sync.Mutex
}

func NewWsController(token string, taskAllocator *queue.TaskAllocator, duration int, mutex *sync.Mutex, doneJob chan string, clients *cache.ClientCache, results, waitingChannels *cache.Cache, readableRequest chan *model.RequestData) *WsController {
	return &WsController{
		clients: clients,
		results: results,
		doneJob: doneJob,
		taskAllocator: taskAllocator,
		waitingChannels: waitingChannels,
		readableRequest: readableRequest,
		mutex: mutex,
		token: token,
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024 * 1000,
	WriteBufferSize: 1024 * 1000,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}
