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

			case <-poller.shutdownDevice:

				poller.shutdownDevice <- true

				return

			case newDevices := <-deviceChannel:

				poller.devicesLock.Lock()

				counters := GetAllCounters()

				newDeviceMap := make(map[string]Device)

				for _, device := range newDevices {

					newDeviceMap[device.IP] = device
				}

				for counterID := range counters {

					updatedDevices := make([]Device, 0)

					// check and add for an existing device

					for _, device := range poller.devices[counterID] {

						if newDevice, found := newDeviceMap[device.IP]; found && !newDevice.IsProvisioned {

							continue
						}

						updatedDevices = append(updatedDevices, device)

					}

					// check and add for a new device

					for _, device := range newDeviceMap {

						if !device.IsProvisioned {

							continue
						}

						updatedDevices = append(updatedDevices, device)
					}

					poller.devices[counterID] = updatedDevices
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
