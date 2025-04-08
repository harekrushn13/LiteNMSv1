package polling

import (
	"fmt"
	"math/rand"
	. "reportdb/config"
	"time"
)

type PollerEngine struct {
	PollInterval time.Duration

	BatchTime time.Duration

	StopTime time.Duration
}

func NewPollerEngine() *PollerEngine {

	return &PollerEngine{
		PollInterval: 1 * time.Second,

		BatchTime: 2500 * time.Millisecond,

		StopTime: 10 * time.Second,
	}
}

func (poller *PollerEngine) PollData(gobalCfg *GlobalConfig) <-chan []RowData {

	out := make(chan []RowData)

	go func() {

		ticker := time.NewTicker(poller.PollInterval)

		defer ticker.Stop()

		batchTicker := time.NewTicker(poller.BatchTime)

		defer batchTicker.Stop()

		stopTimer := time.NewTimer(poller.StopTime)

		defer stopTimer.Stop()

		var batch []RowData

		for {

			select {

			case <-ticker.C:

				t := time.Now().Unix()

				for objectID := uint32(1); objectID <= gobalCfg.ObjectCount; objectID++ {

					for counterID := uint16(1); counterID <= gobalCfg.CounterCount; counterID++ {

						var value interface{}

						switch CounterTypeMapping[counterID] {

						case TypeUint64:

							value = rand.Uint64()

						case TypeFloat64:

							value = rand.Float64() * 10

						case TypeString:

							value = generateRandomString(rand.Intn(50) + 1)
						}

						batch = append(batch, RowData{

							ObjectId: objectID,

							CounterId: counterID,

							Timestamp: uint32(t),

							Value: value,
						})

						if counterID == 2 && objectID == 3 {

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
