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

			index := uint8((query.ObjectId + uint32(query.CounterId)) % uint32(numReaders))

			if index >= numReaders || index < 0 {

				log.Printf("DistributeQuery error : Query index is out of range")

				continue
			}

			readers[index].queryEvents <- query
		}
	}()
}
