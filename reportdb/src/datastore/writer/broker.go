package writer

import (
	"log"
	. "reportdb/utils"
	"sync"
)

func DistributeData(dataChannel chan []Events, writers []*Writer, waitGroup *sync.WaitGroup) {

	waitGroup.Add(1)

	go func(dataChannel chan []Events, writers []*Writer, waitGroup *sync.WaitGroup) {

		defer waitGroup.Done()

		for batch := range dataChannel {

			for _, row := range batch {

				numWriters, err := GetWriters()

				if err != nil {

					log.Fatal("distributeData error :", err)
				}

				index := uint8((uint32(row.CounterId) + row.ObjectId) % uint32(numWriters))

				if index >= numWriters || index < 0 {

					log.Fatal("distributeData error : Writer index is out of range")
				}

				writers[index].events <- row
			}
		}

		for _, writer := range writers {

			close(writer.events)
		}

		return

	}(dataChannel, writers, waitGroup)

}
