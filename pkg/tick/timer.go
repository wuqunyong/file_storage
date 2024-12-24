package tick

import (
	"container/heap"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type TimerCb func(uint64)

// An Item is something we manage in a priority queue.
type Item struct {
	expireTime int64 // The priority of the item in the queue.
	// The index is needed by update and is maintained by the heap.Interface methods.
	index int // The index of the item in the heap.

	id        uint64
	when      time.Time
	period    int64
	oneshot   bool
	task      TimerCb
	callTimes int64
}

func (item *Item) GetId() uint64 {
	return item.id
}

func (item *Item) GetExpireTime() int64 {
	return item.expireTime
}

func (item *Item) GetOneshot() bool {
	return item.oneshot
}

func (item *Item) SetOneshot(value bool) {
	item.oneshot = value
}

func (item *Item) Restore() {
	item.callTimes++
	item.expireTime = item.expireTime + item.period
}

func (item *Item) Run() {
	if item.task == nil {
		return
	}
	item.task(item.id)
}

func GetTimeAfterInterval(milliSec int64) int64 {
	iCur := time.Now().UnixMilli()
	iExpireTime := iCur + milliSec
	return iExpireTime
}

func NewItem(period int64, task TimerCb) *Item {
	expireTime := GetTimeAfterInterval(period)
	return &Item{
		when:       time.Now(),
		period:     period,
		expireTime: expireTime,
		oneshot:    true,
		task:       task,
		callTimes:  0,
	}
}

// A PriorityQueue implements heap.Interface and holds Items.
type PriorityQueue []*Item

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	return pq[i].expireTime < pq[j].expireTime
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

type TimerQueue struct {
	lock   sync.Mutex
	id     uint64
	maxId  uint64
	queue  PriorityQueue
	idItem map[uint64]*Item
}

func NewTimerQueue() *TimerQueue {
	timer := &TimerQueue{
		id:     0,
		queue:  make(PriorityQueue, 0),
		idItem: make(map[uint64]*Item),
	}
	heap.Init(&timer.queue)

	return timer
}

func (t *TimerQueue) GenId() uint64 {
	return atomic.AddUint64(&t.id, 1)
}

func (t *TimerQueue) Len() int {
	t.lock.Lock()
	defer t.lock.Unlock()

	return t.queue.Len()
}

func (t *TimerQueue) Push(item *Item) {
	t.lock.Lock()
	defer t.lock.Unlock()

	id := t.GenId()
	if id > t.maxId {
		t.maxId = id
	}
	item.id = id
	t.idItem[id] = item
	heap.Push(&t.queue, item)
}

func (t *TimerQueue) Restore(item *Item) error {
	t.lock.Lock()
	defer t.lock.Unlock()

	id := item.id
	_, ok := t.idItem[id]
	if ok {
		fmt.Println("Restore duplicate id:", id)
		return errors.New("Restore duplicate id:" + strconv.FormatUint(id, 10))
	}

	if id > t.maxId {
		fmt.Println("Restore overflow id:", id)
		return errors.New("Restore overflow id:" + strconv.FormatUint(id, 10))
	}

	t.idItem[id] = item
	heap.Push(&t.queue, item)

	return nil
}

func (t *TimerQueue) Pop() *Item {
	t.lock.Lock()
	defer t.lock.Unlock()

	if t.queue.Len() <= 0 {
		return nil
	}

	item := heap.Pop(&t.queue).(*Item)
	delete(t.idItem, item.id)
	return item
}

func (t *TimerQueue) Peek() *Item {
	t.lock.Lock()
	defer t.lock.Unlock()

	if t.queue.Len() <= 0 {
		return nil
	}

	return t.queue[0]
}

func (t *TimerQueue) Remove(id uint64) *Item {
	t.lock.Lock()
	defer t.lock.Unlock()

	item, ok := t.idItem[id]
	if !ok {
		return nil
	}

	remItem := heap.Remove(&t.queue, item.index).(*Item)
	delete(t.idItem, item.id)
	return remItem
}
