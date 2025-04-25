package writer

import (
	"log"
	. "reportdb/utils"
)

func DistributeData(dataChannel chan []Events, writers []*Writer) {

	go func() {

		defer ShutdownWriters(writers)

		numWriters := uint8(len(writers))

		for batch := range dataChannel {

			for _, row := range batch {

				index := uint8((uint32(row.CounterId) + row.ObjectId) % uint32(numWriters))

				if index >= numWriters || index < 0 {

					log.Printf("distributeData error : Writer index is out of range")

					continue
				}

				writers[index].events <- row
			}
		}

		return

	}()

	return
}
