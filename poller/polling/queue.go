package polling

import "time"

type Task struct {
	CounterID uint16

	NextExecution time.Time

	Interval time.Duration

	Index int
}

type PriorityQueue []*Task

func (pq *PriorityQueue) Len() int {

	return len(*pq)
}

func (pq *PriorityQueue) Less(i, j int) bool {

	return (*pq)[i].NextExecution.Before((*pq)[j].NextExecution)
}

func (pq *PriorityQueue) Swap(i, j int) {

	(*pq)[i], (*pq)[j] = (*pq)[j], (*pq)[i]

	(*pq)[i].Index = i

	(*pq)[j].Index = j
}

func (pq *PriorityQueue) Push(x interface{}) {

	n := len(*pq)

	item := x.(*Task)

	item.Index = n

	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() interface{} {

	old := *pq

	n := len(old)

	item := old[n-1]

	old[n-1] = nil

	item.Index = -1

	*pq = old[0 : n-1]

	return item
}
