package services

import (
	"context"
	"errors"
	"fmt"
	"insight-engine-backend/models"
	"sort"
	"sync"
	"time"
)

// QueryPriority defines the priority level of a query
type QueryPriority int

const (
	PriorityCritical QueryPriority = 0
	PriorityHigh     QueryPriority = 1
	PriorityNormal   QueryPriority = 2
	PriorityLow      QueryPriority = 3
)

// QueryJob represents a query execution request
type QueryJob struct {
	ID            string
	Conn          *models.Connection
	Query         string
	Params        []interface{} // Added params support
	Limit         *int
	Offset        *int
	Priority      QueryPriority
	SubmittedAt   time.Time
	ResultChannel chan QueryResultWrapper
	Ctx           context.Context
}

// QueryResultWrapper wraps the result or error
type QueryResultWrapper struct {
	Result *models.QueryResult
	Error  error
}

// QueryQueueService manages query execution queuing and resource allocation
type QueryQueueService struct {
	executor      *QueryExecutor
	jobQueue      []*QueryJob
	queueLock     sync.Mutex
	semaphore     chan struct{}
	maxConcurrent int
	shutdown      chan struct{}
}

// NewQueryQueueService creates a new query queue manager
func NewQueryQueueService(executor *QueryExecutor, maxConcurrent int) *QueryQueueService {
	if maxConcurrent <= 0 {
		maxConcurrent = 10 // Default to 10 concurrent queries
	}

	qs := &QueryQueueService{
		executor:      executor,
		jobQueue:      make([]*QueryJob, 0),
		semaphore:     make(chan struct{}, maxConcurrent),
		maxConcurrent: maxConcurrent,
		shutdown:      make(chan struct{}),
	}

	go qs.workerLoop()

	return qs
}

// Enqueue adds a query to the queue and waits for the result
func (qs *QueryQueueService) Enqueue(ctx context.Context, conn *models.Connection, query string, params []interface{}, limit, offset *int, priority QueryPriority) (*models.QueryResult, error) {
	resultChan := make(chan QueryResultWrapper, 1) // Buffered channel to prevent blocking worker

	job := &QueryJob{
		ID:            fmt.Sprintf("job-%d", time.Now().UnixNano()),
		Conn:          conn,
		Query:         query,
		Params:        params,
		Limit:         limit,
		Offset:        offset,
		Priority:      priority,
		SubmittedAt:   time.Now(),
		ResultChannel: resultChan,
		Ctx:           ctx,
	}

	qs.addJob(job)

	// Wait for result or context cancellation
	select {
	case wrapper := <-resultChan:
		return wrapper.Result, wrapper.Error
	case <-ctx.Done():
		// Note: The job might still be in queue or running.
		// If implementation supports cancellation, we should remove it from queue.
		// For now, we just return error to caller.
		return nil, ctx.Err()
	case <-qs.shutdown:
		return nil, errors.New("service shutting down")
	}
}

// addJob safely adds a job to the queue and sorts by priority
func (qs *QueryQueueService) addJob(job *QueryJob) {
	qs.queueLock.Lock()
	defer qs.queueLock.Unlock()

	qs.jobQueue = append(qs.jobQueue, job)

	// Sort by Priority (Ascending: 0 is highest) and then SubmittedAt
	sort.SliceStable(qs.jobQueue, func(i, j int) bool {
		if qs.jobQueue[i].Priority != qs.jobQueue[j].Priority {
			return qs.jobQueue[i].Priority < qs.jobQueue[j].Priority
		}
		return qs.jobQueue[i].SubmittedAt.Before(qs.jobQueue[j].SubmittedAt)
	})

	// Trigger worker if needed?
	// The workerLoop polls deeply, or we could use a condition variable.
	// For simplicity with select/default in worker, polling or signal chan is fine.
	// But pure polling is CPU heavy. Better to use a signal.
}

// workerLoop constantly attempts to process jobs
func (qs *QueryQueueService) workerLoop() {
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-qs.shutdown:
			return
		case <-ticker.C:
			qs.processNextJob()
		}
	}
}

// processNextJob picks the highest priority job and runs it
func (qs *QueryQueueService) processNextJob() {
	// 1. Try to acquire semaphore (limit concurrency)
	select {
	case qs.semaphore <- struct{}{}:
		// Acquired execution slot
	default:
		// Max concurrent reached, wait
		return
	}

	// 2. Get next job
	job := qs.popJob()
	if job == nil {
		// No jobs, release semaphore
		<-qs.semaphore
		return
	}

	// 3. Execute in background (goroutine)
	go func(j *QueryJob) {
		defer func() {
			// Release semaphore when done
			<-qs.semaphore
		}()

		// Check if context is already cancelled
		if j.Ctx.Err() != nil {
			j.ResultChannel <- QueryResultWrapper{Error: j.Ctx.Err()}
			return
		}

		// Execute
		result, err := qs.executor.Execute(j.Ctx, j.Conn, j.Query, j.Params, j.Limit, j.Offset)

		// Send result (non-blocking if possible, but channel is buffered 1)
		// We use a select to avoid leaking if receiver is gone (though Enqueue handles that)
		select {
		case j.ResultChannel <- QueryResultWrapper{Result: result, Error: err}:
		default:
			// Receiver likely gave up (context timeout)
		}
		close(j.ResultChannel)
	}(job)
}

// popJob safely removes the first job from the queue
func (qs *QueryQueueService) popJob() *QueryJob {
	qs.queueLock.Lock()
	defer qs.queueLock.Unlock()

	if len(qs.jobQueue) == 0 {
		return nil
	}

	job := qs.jobQueue[0]
	qs.jobQueue = qs.jobQueue[1:]
	return job
}

// Shutdown stops the queue service
func (qs *QueryQueueService) Shutdown() {
	close(qs.shutdown)
}
