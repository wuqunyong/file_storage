package tick

import (
	"fmt"
	"testing"
	"time"
)

func Test(t *testing.T) {
	timerQueue := NewTimerQueue()

	items := map[string]int64{
		"banana": 3, "apple": 2, "pear": 4,
	}
	for _, priority := range items {
		iValue := time.Duration(priority) * time.Second
		item1 := NewTimer(iValue, func(id uint64) {
			fmt.Println("Id:", id, "priority:", iValue)
		})
		timerQueue.Push(item1)
	}

	item1 := NewTimer(10, func(id uint64) {
		fmt.Println("Id:", id, "priority:", 10)
	})
	timerQueue.Push(item1)
	iId := item1.GetId()

	item2 := NewTimer(20, func(id uint64) {
		fmt.Println("Id:", id, "priority:", 20)
	})
	timerQueue.Push(item2)

	remItem := timerQueue.Remove(iId)
	remItem.Run()
	fmt.Println("remItem")

	timerQueue.Push(remItem)

	for timerQueue.Len() > 0 {
		item := timerQueue.Pop()
		item.Run()
	}
}
