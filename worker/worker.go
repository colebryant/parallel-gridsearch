package worker

import (
	"math/rand"
	"proj3/queue"
	"sync"
)

// SharedContext contains shared variables for goroutines
type SharedContext struct {
	Mutex      *sync.Mutex
	Group      *sync.WaitGroup
	TrainAccuracies []float64
	ParamSchemes [][]string
}

// Worker contains fields need for goroutine to execute in parallel scheme
type Worker struct {
	queueID int
	myQueue queue.Queue
	queues  []queue.Queue
	ctx     interface{}
	done chan bool
}


// NewWorker creates and returns a new Worker
func NewWorker(assignedQueue int, ctx interface{}, queues []queue.Queue) *Worker {
	myQueue := queues[assignedQueue]
	done := make(chan bool)
	return &Worker{queueID: assignedQueue, myQueue: myQueue, queues: queues, ctx: ctx, done: done}
}


// Run initiates the worker to begin taking tasks from its queue / stealing from other queues
func (worker *Worker) Run() {

	ctx := worker.ctx.(*SharedContext)
	ctx.Group.Add(1)
	go func() {
		for {
			select {
			case <- worker.done:
				// If done, get any remaining tasks from local queue
				for {
					task := worker.myQueue.Dequeue()
					if task != nil {
						task(worker.ctx)
						// With probability 1 / size of queue, balance. Also, account for edge case with only one queue/thread
						mySize := int(worker.myQueue.GetSize())
						if rand.Intn(mySize + 1) == mySize && len(worker.queues) != 1 {
							worker.balance()
						}
					} else {
						// Exit if nothing remaining in queue
						ctx.Group.Done()
						return
					}
				}
			default:
				// Default behavior: attempt to get task from local queue
				task := worker.myQueue.Dequeue()
				if task != nil {
					task(worker.ctx)
				}
				// With probability 1 / size of queue, balance. Also, account for edge case with only one queue/thread
				mySize := int(worker.myQueue.GetSize())
				if rand.Intn(mySize + 1) == mySize && len(worker.queues) != 1 {
					worker.balance()
				}
			}
		}
	}()
}


// balance performs work balancing between the worker's queue and another random queue
func (worker *Worker) balance() {
	// Get other queues in slice
	var otherQueues []int
	for i := 0; i < len(worker.queues); i++ {
		if i != worker.queueID {
			otherQueues = append(otherQueues, i)
		}
	}
	// Pick random queue (not including current)
	randNum := rand.Intn(len(otherQueues))
	randIndex := otherQueues[randNum]
	randQ := (worker.queues[randIndex]).(*queue.CoarseGrainedQueue)

	// Balance the two queues
	worker.myQueue.Balance(randQ)
}


// Exit notifies the Worker that no more work is to be pushed to the queue, and can end
// when no more work can be retrieved
func (worker *Worker) Exit() {
	worker.done <- true
}