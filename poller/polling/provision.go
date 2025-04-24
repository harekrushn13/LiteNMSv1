package polling

import (
	"container/heap"
	. "poller/utils"
	"time"
)

func (poller *Poller) SetProvisionedDevices(deviceChannel chan []Device) {

	go func() {

		for {

			select {

			case newDevices := <-deviceChannel:

				poller.devicesLock.Lock()

				counters := GetAllCounters()

				for counterID, _ := range counters {

					for _, newDevice := range newDevices {

						exists := false

						for _, existingDevice := range poller.devices[counterID] {

							if existingDevice == newDevice {

								exists = true

								break
							}
						}

						if !exists {

							poller.devices[counterID] = append(poller.devices[counterID], newDevice)
						}
					}
				}

				poller.devicesLock.Unlock()

				if len(newDevices) > 0 {

					poller.makeTaskQueue()
				}
			}
		}

	}()
}

func (poller *Poller) makeTaskQueue() {

	poller.taskLock.Lock()

	defer poller.taskLock.Unlock()

	poller.taskQueue = poller.taskQueue[:0]

	now := time.Now()

	for counterID := range poller.devices {

		if len(poller.devices[counterID]) > 0 {

			interval := time.Duration(GetCounterPollingInterval(counterID)) * time.Second

			if interval > 0 {

				heap.Push(&poller.taskQueue, &Task{

					CounterID: counterID,

					NextExecution: now.Add(interval),

					Interval: interval,
				})
			}
		}
	}
}
