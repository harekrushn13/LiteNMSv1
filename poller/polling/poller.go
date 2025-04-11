package polling

import (
	"fmt"
	"math/rand"
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

func PollData(waitGroup *sync.WaitGroup) <-chan []Events {

	out := make(chan []Events)

	waitGroup.Add(1)

	go func() {

		defer waitGroup.Done()

		time.Sleep(5 * time.Second)

		ticker := time.NewTicker(time.Duration(GetPollingInterval()) * time.Second)

		defer ticker.Stop()

		batchTicker := time.NewTicker(time.Duration(GetBatchInterval()) * time.Millisecond)

		defer batchTicker.Stop()

		stopTimer := time.NewTimer(time.Duration(GetStopTime()) * time.Millisecond)

		defer stopTimer.Stop()

		var batch []Events

		for {

			select {

			case <-ticker.C:

				timestamp := time.Now().Unix()

				for objectId := uint32(1); objectId <= GetObjects(); objectId++ {

					for counterId := uint16(1); counterId <= GetCounters(); counterId++ {

						var value interface{}

						switch GetCounterType(counterId) {

						case TypeUint64:

							value = rand.Uint64()

						case TypeFloat64:

							value = rand.Float64() * 10

						case TypeString:

							value = generateRandomString(rand.Intn(50) + 1)
						}

						batch = append(batch, Events{

							ObjectId: objectId,

							CounterId: counterId,

							Timestamp: uint32(timestamp),

							Value: value,
						})

						if counterId == 3 && objectId == 3 {

							fmt.Println("data : ", batch[len(batch)-1])
						}
					}
				}

			case <-batchTicker.C:

				if len(batch) > 0 {

					fmt.Println("batch send")

					out <- batch

					batch = nil
				}

			case <-stopTimer.C:

				if len(batch) > 0 {

					fmt.Println("poll stop")

					out <- batch
				}

				close(out)

				return
			}

		}
	}()

	return out
}

func generateRandomString(length int) string {

	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	b := make([]byte, length)

	for i := range b {

		b[i] = charset[rand.Intn(len(charset))]
	}

	return string(b)
}
