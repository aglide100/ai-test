package queue

import (
	"sync"
	"time"

	"github.com/aglide100/ai-test/pkg/logger"
	"github.com/aglide100/ai-test/pkg/model"
	"go.uber.org/zap"
)

type JobQueue struct {
	data *PriorityQueue
	mu *sync.Mutex
}

type Allocate struct {
	Who model.Runner
	Job *model.Job
}

func NewJobQueue(mu *sync.Mutex) *JobQueue {
	return &JobQueue{
		data: NewPriorityQueue(mu),
		mu: mu,
	}
}

func (q *JobQueue) Len() int {
	return q.data.Len()
}

func (q *JobQueue) RemoveByClientUUID(uuid string) (bool, *model.Job) {
	q.mu.Lock()
	defer q.mu.Unlock()

	tmp := make([]*QueueItem, len(q.data.queue))
	copy(tmp, q.data.queue)

	Index := -1
	found := false
	res := &Allocate{}

	for idx, val := range tmp {
		allocate := val.Value.(*Allocate)

		if allocate.Who.Who == uuid {
			Index = idx 
			found = true
			tmp[idx].Index--
			res = allocate
			break
		}
	}

	if !found {
		logger.Debug("Can't find item")
		return false, nil
	}

	if len(tmp) == 1 {
		q.data.queue = make([]*QueueItem, 0)
		return true, res.Job
	} 
	
	if Index == len(tmp)-1 {
		q.data.queue = tmp[:Index]
		return true, res.Job
	}
		
	q.data.queue = append(tmp[:Index], tmp[Index+1])
	return true, res.Job
}

func (q *JobQueue) RemoveByJobId(jobId string) (bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	tmp := make([]*QueueItem, len(q.data.queue))
	copy(tmp, q.data.queue)

	Index := -1
	found := false

	for idx, val := range tmp {
		allocate := val.Value.(*Allocate)
		if allocate.Job.ID == jobId {
			Index = idx 
			found = true
			continue
		}

		if found {
			tmp[idx].Index--
		}
	}

	if !found {
		logger.Debug("Can't find item")
		return false
	}

	if len(tmp) == 1 {
		q.data.queue = make([]*QueueItem, 0)
		return true
	} 
	
	if Index == len(tmp)-1 {
		q.data.queue = tmp[:Index]
		return true
	}
		
	q.data.queue = append(tmp[:Index], tmp[Index+1])
	return true
}

func (q *JobQueue) CleanUp(t int) []*QueueItem {
    currentTime := time.Now()
    var timeOutItems []*QueueItem

	q.mu.Lock()
	for _, val := range q.data.queue {
		allocate := val.Value.(*Allocate)
		if val.When.IsZero() {
			logger.Error("time is weird", zap.Any("tmp", val))
		} else {
			duration := currentTime.Sub(val.When)
			if duration >= time.Second * time.Duration(t) {
				logger.Info("Timeout!", zap.Any("prompt", allocate.Job.PromptA), zap.Any("by", allocate.Who))
				timeOutItems = append(timeOutItems, val)
			}
		}
	}
	q.mu.Unlock()

	for _, val := range timeOutItems {
		allocate := val.Value.(*Allocate)

		q.Remove(allocate.Job)
	}

    return timeOutItems
}

func (q *JobQueue) Remove(item *model.Job) (bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	tmp := make([]*QueueItem, len(q.data.queue))
	copy(tmp, q.data.queue)

	Index := -1
	found := false

	for idx, val := range tmp {
		allocate := val.Value.(*Allocate)
		if allocate.Job.ID == item.ID {
			Index = idx
			found = true
			continue
		}

		if found {
			tmp[idx].Index--
		}
	}

	if !found {
		logger.Debug("Can't find item")
		return false
	}

	if len(tmp) == 1 {
		tmp = make([]*QueueItem, 0)
		q.data.queue = tmp
		return true
	}

	if Index == len(tmp)-1 {
		q.data.queue = tmp[:Index]
		return true
	}

	q.data.queue = append(tmp[:Index], tmp[Index+1])
	return true
}

func (q *JobQueue) Push(item *QueueItem) {
	q.data.Push(item)
}

func (q *JobQueue) Pop() (*QueueItem, bool) {
	return q.data.Pop()
}

func (q *JobQueue) Get(clientUUID string) (*Allocate, bool) {
	for _, val := range q.data.queue {
		allocate := val.Value.(*Allocate)
		if (allocate.Who.Who == clientUUID) {
			return allocate, true
		}
	}

	return nil, false
}