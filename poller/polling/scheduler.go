package polling

import (
	"container/heap"
	"fmt"
	"golang.org/x/crypto/ssh"
	. "poller/utils"
	"time"
)

func InitScheduler(batch *[]Events) {

	pq := make(PriorityQueue, 0)

	taskQueue = &pq

	heap.Init(taskQueue)

	workerChan = make(chan *Task, 10)

	shutdownChan = make(chan struct{})

	go startScheduler()

	for i := 0; i < 5; i++ { // 5 workers

		go startWorker(batch)
	}
}

func startScheduler() {

	for {

		select {

		case <-shutdownChan:

			return

		default:

			now := time.Now()

			taskLock.Lock()

			if taskQueue.Len() > 0 {

				nextTask := (*taskQueue)[0]

				if nextTask.NextExecution.Before(now) || nextTask.NextExecution.Equal(now) {

					task := heap.Pop(taskQueue).(*Task)

					workerChan <- task

					task.NextExecution = now.Add(task.Interval)

					heap.Push(taskQueue, task)
				}
			}

			taskLock.Unlock()
		}
	}
}

func startWorker(batch *[]Events) {

	for {

		select {

		case <-shutdownChan:

			return

		case task := <-workerChan:

			pollCounter(task.CounterID, batch)
		}
	}
}

func pollCounter(counterID uint16, batch *[]Events) {

	devicesLock.RLock()

	defer devicesLock.RUnlock()

	deviceList, exists := devices[counterID]

	if !exists || len(deviceList) == 0 {

		return
	}

	timestamp := uint32(time.Now().Unix())

	for _, device := range deviceList {

		value, err := getCounterViaSSH(device, counterID)

		if err != nil {

			fmt.Printf("Error polling device %s (ID: %d) for counter %d: %v\n",
				device.IP, device.ObjectID, counterID, err)

			continue
		}

		*batch = append(*batch, Events{

			ObjectId: device.ObjectID,

			CounterId: counterID,

			Timestamp: timestamp,

			Value: value,
		})
	}

}

func getCounterViaSSH(device Device, counterId uint16) (interface{}, error) {

	config := &ssh.ClientConfig{

		User: device.Username,

		Auth: []ssh.AuthMethod{
			ssh.Password(device.Password),
		},

		HostKeyCallback: ssh.InsecureIgnoreHostKey(),

		Timeout: 5 * time.Second,
	}

	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", device.IP, device.Port), config)

	if err != nil {

		return nil, err
	}

	defer client.Close()

	session, err := client.NewSession()

	if err != nil {

		return nil, err
	}

	defer session.Close()

	var cmd string

	switch counterId {

	case 1:
		cmd = `free -b | awk '/Mem:/ {print $3}'`

	case 2:
		cmd = `top -bn1 | awk '/%Cpu/ {print 100 - $8}'`

	case 3:
		cmd = `hostname`

	default:
		return nil, fmt.Errorf("unknown counter ID: %d", counterId)
	}

	output, err := session.Output(cmd)

	if err != nil {

		return nil, err
	}

	switch GetCounterType(counterId) {

	case TypeUint64:

		var uintValue uint64

		_, err := fmt.Sscanf(string(output), "%d", &uintValue)

		return uintValue, err

	case TypeFloat64:

		var floatValue float64

		_, err := fmt.Sscanf(string(output), "%f", &floatValue)

		return floatValue, err

	case TypeString:

		return string(output), nil

	default:

		return nil, fmt.Errorf("unsupported counter type")
	}
}
