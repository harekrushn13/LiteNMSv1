package polling

import (
	"fmt"
	"golang.org/x/crypto/ssh"
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
	devices []Device

	devicesLock sync.RWMutex
)

func SetProvisionedDevices(newDevices []Device) {

	devicesLock.Lock()

	defer devicesLock.Unlock()

	devices = newDevices
}

func PollData(waitGroup *sync.WaitGroup) <-chan []Events {

	out := make(chan []Events, 10)

	waitGroup.Add(1)

	go func() {

		defer waitGroup.Done()

		ticker1s := time.NewTicker(1 * time.Second)

		ticker2s := time.NewTicker(2 * time.Second)

		ticker5s := time.NewTicker(3 * time.Second)

		batchTicker := time.NewTicker(time.Duration(GetBatchInterval()) * time.Millisecond)

		defer ticker1s.Stop()

		defer ticker2s.Stop()

		defer ticker5s.Stop()

		defer batchTicker.Stop()

		var batch []Events

		for {

			select {

			case <-ticker1s.C:

				pollCounter(1, &batch)

			case <-ticker2s.C:

				pollCounter(2, &batch)

			case <-ticker5s.C:

				pollCounter(3, &batch)

			case <-batchTicker.C:

				if len(batch) > 0 {

					out <- batch

					batch = nil
				}
			}
		}
	}()

	return out
}

func pollCounter(counterId uint16, batch *[]Events) {

	devicesLock.RLock()

	defer devicesLock.RUnlock()

	timestamp := uint32(time.Now().Unix())

	for _, device := range devices {

		value, err := getCounterViaSSH(device, counterId)

		if err != nil {

			fmt.Printf("Error polling device %s (ID: %d) for counter %d: %v\n",
				device.IP, device.ObjectID, counterId, err)

			continue
		}

		*batch = append(*batch, Events{

			ObjectId: device.ObjectID,

			CounterId: counterId,

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
