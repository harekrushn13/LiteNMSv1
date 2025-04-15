package reader

import (
	"log"
	. "reportdb/utils"
)

func DistributeQuery(queryChannel chan Query, readers []*Reader) {

	go func() {

		defer ShutdownReaders(readers)

		numReaders := uint8(len(readers))

		for query := range queryChannel {

			index := uint8(query.RequestID % uint64(numReaders))

			if index >= numReaders || index < 0 {

				log.Printf("distributeData error : Writer index is out of range")

				return
			}

			readers[index].queryEvents <- query
		}
	}()
}
