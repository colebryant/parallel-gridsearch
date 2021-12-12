package queue

import (
	"sync"
	"sync/atomic"
)

// heuristic threshold value for determining if should re-balance queues
const thresholdSizeDiff = 2

type Runnable func(arg interface{})

// Queue is the interface for a queue data structure (that is also capable of work balancing)
type Queue interface {
	Enqueue(task Runnable)
	Dequeue() Runnable
	IsEmpty() bool //returns whether the queue is empty
	GetSize() uint64
	Balance(queue * CoarseGrainedQueue)
}

// CoarseGrainedQueue is a lock-based concurrent queue (unbounded)
type CoarseGrainedQueue struct {
	head *node
	tail *node
	size uint64
	enqLock *sync.Mutex
	deqLock *sync.Mutex
}

// NewCoarseGrainedQueue creates and returns a new CoarseGrainedQueue
func NewCoarseGrainedQueue() Queue {
	sentinel := newNode(nil, nil)
	size := uint64(0)
	enqLock := sync.Mutex{}
	deqLock := sync.Mutex{}
	return &CoarseGrainedQueue{head: sentinel, tail: sentinel, size: size, enqLock: &enqLock, deqLock: &deqLock}
}

// node functions as a part of the CoarseGrainedQueue and holds a Runnable task
type node struct {
	task Runnable
	next *node
}

// newNode creates and returns a new node
func newNode(task Runnable, next *node) *node {
	return &node{task, next}
}

// Enqueue is used by the thread to add a task to the queue
func (cgq *CoarseGrainedQueue) Enqueue(task Runnable) {
	cgq.enqLock.Lock()
	defer cgq.enqLock.Unlock()
	enqNode := newNode(task, nil)
	cgq.tail.next = enqNode
	cgq.tail = enqNode
	atomic.AddUint64(&cgq.size, 1)
}

// Dequeue is used by a thread to take task from the head of dequeue
func (cgq *CoarseGrainedQueue) Dequeue() Runnable {
	cgq.deqLock.Lock()
	defer cgq.deqLock.Unlock()
	if cgq.IsEmpty() {
		return nil
	}
	task := cgq.head.next.task
	cgq.head = cgq.head.next
	atomic.AddUint64(&cgq.size, ^uint64(0))
	return task
}

// IsEmpty checks to see if the queue is empty
func (cgq *CoarseGrainedQueue) IsEmpty() bool {
	return atomic.LoadUint64(&cgq.size) == 0
}

// GetSize returns the size of the queue
func (cgq *CoarseGrainedQueue) GetSize() uint64 {
	return atomic.LoadUint64(&cgq.size)
}

// Balance balances the current queue with another queue
func (cgq *CoarseGrainedQueue) Balance(otherQueue *CoarseGrainedQueue) {
	mySize := cgq.GetSize()
	otherSize := otherQueue.GetSize()

	// Determine max and min queue sizes
	var max, min uint64
	var maxQueue, minQueue *CoarseGrainedQueue
	if mySize > otherSize {
		max = mySize
		maxQueue = cgq
		min = otherSize
		minQueue = otherQueue
	} else {
		max = otherSize
		maxQueue = otherQueue
		min = mySize
		minQueue = cgq
	}

	// Don't try and balance if size diff is less than threshold heuristic
	if max - min < thresholdSizeDiff {
		return
	}

	// Lock down both queues
	minQueue.enqLock.Lock()
	minQueue.deqLock.Lock()
	maxQueue.enqLock.Lock()
	maxQueue.deqLock.Lock()

	// While max queue size is greater than min queue size, keep dequeuing task from max queue and enqueuing in min queue
	for max > min {
		task := maxQueue.head.next.task
		maxQueue.head = maxQueue.head.next
		maxQueue.size--
		max--

		enqNode := newNode(task, nil)
		minQueue.tail.next = enqNode
		minQueue.tail = enqNode
		minQueue.size++
		min++
	}
	// Unlock both queues
	minQueue.enqLock.Unlock()
	minQueue.deqLock.Unlock()
	maxQueue.enqLock.Unlock()
	maxQueue.deqLock.Unlock()
}

