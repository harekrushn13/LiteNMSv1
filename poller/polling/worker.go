package polling

import (
	"fmt"
	"go.uber.org/zap"
	"golang.org/x/crypto/ssh"
	. "poller/logger"
	. "poller/utils"
	"strings"
	"sync"
	"time"
)

func (poller *Poller) startWorker(eventChannel chan Events) {

	pollDevicePool := make(chan struct{}, GetPollDeviceBuffer())

	for i := 0; i < GetPollDeviceBuffer(); i++ {

		pollDevicePool <- struct{}{}
	}

	for {

		select {

		case <-poller.shutdownWorker:

			close(poller.workerChan)

			poller.shutdownWorker <- true

			return

		case task := <-poller.workerChan:

			if task == nil {

				continue
			}

			poller.pollCounter(task.CounterID, eventChannel, pollDevicePool)
		}
	}
}

func (poller *Poller) pollCounter(counterID uint16, eventChannel chan Events, pollDevicePool chan struct{}) {

	poller.devicesLock.RLock()

	defer poller.devicesLock.RUnlock()

	deviceList, exists := poller.devices[counterID]

	if !exists || len(deviceList) == 0 {

		return
	}

	timestamp := uint32(time.Now().Unix())

	var wg sync.WaitGroup

	for _, device := range deviceList {

		<-pollDevicePool

		wg.Add(1)

		go func() {

			defer wg.Done()

			defer func() {

				pollDevicePool <- struct{}{}
			}()

			data, err := fetchDataViaSSH(device, counterID)

			if err != nil {

				Logger.Info("Error polling device",
					zap.String("ip", device.IP),
					zap.Uint32("objectID", device.ObjectID),
					zap.Uint16("counterID", counterID),
					zap.Error(err))

				return
			}

			eventChannel <- Events{

				ObjectId: device.ObjectID,

				CounterId: counterID,

				Timestamp: timestamp,

				Value: data,
			}

		}()
	}

	wg.Wait()

}

func fetchDataViaSSH(device Device, counterId uint16) (interface{}, error) {

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

		return strings.TrimSpace(string(output)), nil

	default:

		return nil, fmt.Errorf("unsupported counter type")
	}

}
