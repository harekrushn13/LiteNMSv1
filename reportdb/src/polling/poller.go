package polling

import (
	"fmt"
	"math/rand"
	. "reportdb/utils"
	"sync"
	"time"
)

func PollData(waitGroup *sync.WaitGroup) <-chan []RowData {

	out := make(chan []RowData)

	waitGroup.Add(1)

	go func() {

		defer waitGroup.Done()

		//ticker := time.NewTicker(time.Duration(GetPollingInterval()) * time.Second)
		ticker := time.NewTicker(1 * time.Second)

		defer ticker.Stop()

		//batchTicker := time.NewTicker(time.Duration(GetBatchInterval()) * time.Millisecond)
		batchTicker := time.NewTicker(2500 * time.Millisecond)

		defer batchTicker.Stop()

		//stopTimer := time.NewTimer(time.Duration(GetStopTime()) * time.Second)
		stopTimer := time.NewTimer(60 * time.Second)

		defer stopTimer.Stop()

		var batch []RowData

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

						batch = append(batch, RowData{

							ObjectId: objectId,

							CounterId: counterId,

							Timestamp: uint32(timestamp),

							Value: value,
						})

						if counterId == 3 && objectId == 3 {

							fmt.Println(batch[len(batch)-1])
						}
					}
				}

			case <-batchTicker.C:

				if len(batch) > 0 {

					out <- batch

					batch = nil
				}

			case <-stopTimer.C:

				if len(batch) > 0 {

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
