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

	batchInterval time.Duration

	numWorkers int

	Shutdown
}

type Shutdown struct {
	shutdownWorker chan bool

	shutdownScheduler chan bool

	shutdownBatcher chan bool

	shutdownDevice chan bool
}

func NewPoller() *Poller {

	return &Poller{

		devices: make(map[uint16][]Device),

		taskQueue: make(PriorityQueue, 0),

		workerChan: make(chan *Task, GetWorkBuffer()),

		batchInterval: time.Duration(GetBatchInterval()) * time.Millisecond,

		numWorkers: GetWorkerCount(),

		Shutdown: Shutdown{
			shutdownWorker: make(chan bool, 1),

			shutdownScheduler: make(chan bool, 1),

			shutdownBatcher: make(chan bool, 1),

			shutdownDevice: make(chan bool, 1),
		},
	}
}

func (poller *Poller) StartPolling(dataChannel chan []Events) {

	eventChannel := make(chan Events, GetEventBuffer())

	for i := 0; i < poller.numWorkers; i++ {

		go poller.startWorker(eventChannel)
	}

	heap.Init(&poller.taskQueue)

	go poller.startScheduler()

	go poller.batchEvents(eventChannel, dataChannel)

}

func (poller *Poller) startScheduler() {

	for {

		select {

		case <-poller.shutdownScheduler:

			poller.taskLock.Lock()

			poller.taskQueue = poller.taskQueue[:0]

			poller.taskLock.Unlock()

			poller.shutdownScheduler <- true

			return

		default:

			poller.taskLock.Lock()

			if poller.taskQueue.Len() > 0 {

				nextTask := poller.taskQueue[0]

				now := time.Now()

				if !nextTask.NextExecution.After(now) {

					task := heap.Pop(&poller.taskQueue).(*Task)

					poller.workerChan <- task

					task.NextExecution = now.Add(task.Interval)

					heap.Push(&poller.taskQueue, task)

				}

				waitTime := time.Until(poller.taskQueue[0].NextExecution)

				poller.taskLock.Unlock()

				if waitTime > 0 {

					time.Sleep(waitTime)
				}

			} else {

				poller.taskLock.Unlock()

				time.Sleep(1 * time.Second)
			}

		}
	}
}

func (poller *Poller) batchEvents(eventChannel chan Events, dataChannel chan []Events) {

	var batch []Events

	batchTicker := time.NewTicker(time.Duration(GetBatchInterval()) * time.Millisecond)

	for {

		select {

		case <-poller.shutdownBatcher:

			batchTicker.Stop()

			poller.shutdownBatcher <- true

			return

		case event := <-eventChannel:

			batch = append(batch, event)

		case <-batchTicker.C:

			if len(batch) > 0 {

				dataChannel <- batch

				batch = nil
			}
		}
	}
}

func (poller *Poller) ShutdownPoller() {

	poller.shutdownDevice <- true

	poller.shutdownScheduler <- true

	poller.shutdownWorker <- true

	poller.shutdownBatcher <- true

	<-poller.shutdownDevice

	<-poller.shutdownScheduler

	<-poller.shutdownWorker

	<-poller.shutdownBatcher

}
