package queue

import (
	"container/heap"
	"sync"
	"sync/atomic"
)

// An Item is something we manage in a priority queue.
type Item struct {
	priority uint64 // The priority of the item in the queue.
	// The index is needed by update and is maintained by the heap.Interface methods.
	index int // The index of the item in the heap.

	id   uint64
	Args any
}

// A PriorityQueue implements heap.Interface and holds Items.
type PriorityQueue []*Item

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	return pq[i].priority < pq[j].priority
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *PriorityQueue) Push(x any) {
	n := len(*pq)
	item := x.(*Item)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // avoid memory leak
	item.index = -1 // for safety
	*pq = old[0 : n-1]
	return item
}

type MsgQueue struct {
	lock   sync.Mutex
	id     uint64
	maxId  uint64
	queue  PriorityQueue
	idItem map[uint64]*Item
}

func NewTimerQueue() *MsgQueue {
	timer := &MsgQueue{
		id:     0,
		queue:  make(PriorityQueue, 0),
		idItem: make(map[uint64]*Item),
	}
	heap.Init(&timer.queue)

	return timer
}

func (t *MsgQueue) GenId() uint64 {
	return atomic.AddUint64(&t.id, 1)
}

func (t *MsgQueue) Len() int {
	t.lock.Lock()
	defer t.lock.Unlock()

	return t.queue.Len()
}

func (t *MsgQueue) Push(args any) {
	t.lock.Lock()
	defer t.lock.Unlock()

	id := t.GenId()
	item := &Item{
		priority: id,
		id:       id,
		Args:     args,
	}
	if id > t.maxId {
		t.maxId = id
	}
	t.idItem[id] = item
	heap.Push(&t.queue, item)
}

func (t *MsgQueue) Pop() *Item {
	t.lock.Lock()
	defer t.lock.Unlock()

	if t.queue.Len() <= 0 {
		return nil
	}

	item := heap.Pop(&t.queue).(*Item)
	delete(t.idItem, item.id)
	return item
}
