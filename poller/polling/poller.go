package polling

import (
	"container/heap"
	. "poller/utils"
	"sync"
	"time"
)

type Events struct {
	ObjectId uint32

	CounterId uint16

	Timestamp uint32

	Value interface{}
}
type Device struct {
	ObjectID uint32 `json:"object_id"`

	IP string `json:"ip"`

	CredentialID uint16 `json:"credential_id"`

	DiscoveryID uint16 `json:"discovery_id"`

	Username string `json:"username"`

	Password string `json:"password"`

	Port uint16 `json:"port"`
}

var (
	devices = make(map[uint16][]Device)

	devicesLock sync.RWMutex

	taskQueue *PriorityQueue

	taskLock sync.Mutex

	workerChan chan *Task

	shutdownChan chan struct{}
)

func SetProvisionedDevices(newDevices []Device) {

	devicesLock.Lock()

	defer devicesLock.Unlock()

	devices = make(map[uint16][]Device)

	counters := GetAllCounters()

	for _, newDevice := range newDevices {

		for counterID, _ := range counters {

			exists := false

			for _, existingDevice := range devices[counterID] {

				if existingDevice.ObjectID == newDevice.ObjectID && existingDevice.IP == newDevice.IP && existingDevice.CredentialID == newDevice.CredentialID && existingDevice.DiscoveryID == newDevice.DiscoveryID {

					exists = true

					break
				}
			}

			if !exists {

				devices[counterID] = append(devices[counterID], newDevice)
			}
		}
	}

	if len(newDevices) > 0 {

		createTaskQueue()
	}
}

func createTaskQueue() {

	taskLock.Lock()

	defer taskLock.Unlock()

	*taskQueue = (*taskQueue)[:0]

	now := time.Now()

	for counterID := range devices {

		if len(devices[counterID]) > 0 {

			interval := time.Duration(GetCounterPollingInterval(counterID)) * time.Second

			if interval > 0 {

				heap.Push(taskQueue, &Task{

					CounterID: counterID,

					NextExecution: now.Add(interval),

					Interval: interval,
				})
			}
		}
	}
}

func PollData(waitGroup *sync.WaitGroup) <-chan []Events {

	dataChannel := make(chan []Events, 10)

	waitGroup.Add(1)

	go func() {

		defer waitGroup.Done()

		var batch []Events

		batchTicker := time.NewTicker(time.Duration(GetBatchInterval()) * time.Millisecond)

		defer batchTicker.Stop()

		InitScheduler(&batch)

		for {

			select {

			case <-batchTicker.C:

				if len(batch) > 0 {

					dataChannel <- batch

					batch = nil
				}
			}
		}
	}()

	return dataChannel
}
