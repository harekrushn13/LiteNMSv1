package polling

import (
	"container/heap"
	. "poller/utils"
	"sync"
	"time"
)

type Poller struct {
	devices map[uint16][]Device // map[counterID]->[]Devices

	devicesLock sync.RWMutex

	taskQueue PriorityQueue

	taskLock sync.Mutex

	workerChan chan *Task

	shutdownChan chan struct{}

	batchInterval time.Duration

	numWorkers int
}

func NewPoller() *Poller {

	return &Poller{

		devices: make(map[uint16][]Device),

		taskQueue: make(PriorityQueue, 0),

		workerChan: make(chan *Task, 10),

		shutdownChan: make(chan struct{}),

		batchInterval: time.Duration(GetBatchInterval()) * time.Millisecond,

		numWorkers: GetWorkerCount(),
	}
}

func (poller *Poller) StartPolling(dataChannel chan []Events) {

	eventChannel := make(chan Events, GetEventBuffer())

	for i := 0; i < poller.numWorkers; i++ {

		go poller.startWorker(eventChannel)
	}

	heap.Init(&poller.taskQueue)

	go poller.startScheduler()

	go func() {

		var batch []Events

		batchTicker := time.NewTicker(time.Duration(GetBatchInterval()) * time.Millisecond)

		for {

			select {

			case event := <-eventChannel:

				batch = append(batch, event)

			case <-batchTicker.C:

				if len(batch) > 0 {

					dataChannel <- batch

					batch = nil
				}
			}
		}
	}()

}
