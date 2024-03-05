package queue

import (
	"sync"

	"github.com/aglide100/ai-test/pkg/logger"
	"github.com/aglide100/ai-test/pkg/model"
	"go.uber.org/zap"
)

type TaskAllocator struct {
	// value := *Allocate
	running *JobQueue
	// value := *model.Job
	waiting *JobQueue
}

func NewTaskAllocator(mutex *sync.Mutex) *TaskAllocator {
	running := NewJobQueue(mutex)
	waiting := NewJobQueue(mutex)
	
	return &TaskAllocator{
		running: running,
		waiting: waiting,
	}
}

func (t *TaskAllocator) Check(jobID string) string {
	found := false

	for _, val := range t.running.data.queue {
		job := val.Value.(*Allocate)

		if (job.Job.ID == jobID) {
			found = true
			break
		}
	}

	if (found) {
		return "running"
	}

	for _, val := range t.waiting.data.queue {
		job := val.Value.(*model.Job)

		if (job.ID == jobID) {
			found = true
			break
		}
	}

	if (found) {
		return "waiting"
	}

	return "none"
}

func (t *TaskAllocator) CleanUp(timeOut int) {
	items := t.running.CleanUp(timeOut)

	for _, val := range items {
		job := val.Value.(*model.Job)
		t.waiting.Push(&QueueItem{
			Value: job,
		})
	}
}

func (t *TaskAllocator) AddInWait(item *model.Job) {
	logger.Info("job", zap.Any("job", item))

	t.waiting.Push(
		&QueueItem{
			Value: item,
		},
	)
}

func (t *TaskAllocator) AddInRunning(item *model.Job, clientUUID string) {
	t.running.Push(
		&QueueItem{
			Value: &Allocate{
				Job: item,
				Who: model.Runner{Who: clientUUID, CurrentWork: item.ID},
			},
		},
	)
}

func (t *TaskAllocator) PopInWait() (*model.Job, bool) {
	res, found := t.waiting.Pop()

	if !found {
		return nil, false
	}
	job := res.Value.(*model.Job)

	return job, true
}

func (t *TaskAllocator) RemoveByJobIdInRunning(jobId string) {
	ok := t.running.RemoveByJobId(jobId)
	if !ok {
		logger.Info("Can't remove by job id", zap.Any("jobid", jobId))
		logger.Info("len", zap.Any("len", t.running.Len()))
	}
}

func (t *TaskAllocator) LenRunning() int {
	return t.running.Len()
}

func (t *TaskAllocator) LenWaiting() int {
	return t.waiting.Len()
}

func (t *TaskAllocator) RemoveRunningByClientUUID(clientUUID string) {
	ok, job := t.running.RemoveByClientUUID(clientUUID)
	if !ok {
		// logger.Error("can't remove by clientUUID", zap.Any("uuid", clientUUID))
		return
	}

	t.waiting.Push(&QueueItem{
		Value: job,
	})
}