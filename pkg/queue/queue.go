package queue

import (
	"sync"
	"time"
)

type PriorityQueue struct {
	queue []*QueueItem
	mu    *sync.Mutex
}

type QueueItem struct {
	Value interface{}
	Index int
	When  time.Time
}

func NewPriorityQueue(mu *sync.Mutex) *PriorityQueue {
	h := &PriorityQueue{
		mu: mu,
	}

	return h
}

func (pq *PriorityQueue) Len() int {
	pq.mu.Lock()
	defer pq.mu.Unlock()
	return len(pq.queue)
}

func (pq *PriorityQueue) Less(i, j int) bool {
	return pq.queue[i].When.Before(pq.queue[j].When)
}

func (pq *PriorityQueue) Push(item *QueueItem) {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	n := len(pq.queue)
	item.Index = n
	item.When = time.Now()
	pq.queue = append(pq.queue, item)
	pq.bubbleUp(len(pq.queue) - 1)
}

func (pq *PriorityQueue) Swap(i, j int) {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	pq.queue[i], pq.queue[j] = pq.queue[j], pq.queue[i]
	pq.queue[i].Index = i
	pq.queue[j].Index = j
}

func (pq *PriorityQueue) Pop() (*QueueItem, bool) {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	if len(pq.queue) == 0 {
		return nil, false
	}

	if len(pq.queue) == 1 {
		item := pq.queue[0]
		pq.queue = []*QueueItem{}
		return item, true
	}

	root := pq.queue[0]
	last := pq.queue[len(pq.queue)-1]

	pq.queue[0] = last
	pq.queue = pq.queue[:len(pq.queue)-1]
	pq.bubbleDown(0)

	return root, true
}

func (pq *PriorityQueue) bubbleUp(index int) {
	for index > 0 {
		parentIdx := (index - 1) / 2

		if pq.queue[index].When.After(pq.queue[parentIdx].When) {
			break
		}

		pq.queue[index], pq.queue[parentIdx] = pq.queue[parentIdx], pq.queue[index]
		index = parentIdx
	}
}

func (pq *PriorityQueue) bubbleDown(index int) {
	currentTime := time.Now()

	for {
		leftIdx := index*2 + 1
		rightIdx := index*2 + 2

		min := index

		if leftIdx < len(pq.queue) {
			durationLeftIdx := currentTime.Sub(pq.queue[leftIdx].When)
			durationMinIdx := currentTime.Sub(pq.queue[min].When)

			if durationMinIdx < durationLeftIdx {
				min = leftIdx
			}
		}

		if rightIdx < len(pq.queue) {
			durationRightIdx := currentTime.Sub(pq.queue[rightIdx].When)
			durationMinIdx := currentTime.Sub(pq.queue[min].When)

			if durationMinIdx < durationRightIdx {
				min = rightIdx
			}
		}

		if min == index {
			break
		}

		pq.queue[index], pq.queue[min] = pq.queue[min], pq.queue[index]
		index = min
	}
}
