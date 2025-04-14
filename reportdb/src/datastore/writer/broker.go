package writer

import (
	"log"
	. "reportdb/utils"
)

func DistributeData(dataChannel chan []Events, writers []*Writer) {

	go func() {

		defer ShutdownWriters(writers)

		numWriters, err := GetWriters()

		if err != nil {

			log.Printf("distributeData error : %v", err)

			return
		}

		for batch := range dataChannel {

			for _, row := range batch {

				index := uint8((uint32(row.CounterId) + row.ObjectId) % uint32(numWriters))

				if index >= numWriters || index < 0 {

					log.Printf("distributeData error : Writer index is out of range")

					return
				}

				writers[index].events <- row
			}
		}

		return

	}()

	return
}
