package reader

import (
	"go.uber.org/zap"
	. "reportdb/logger"
	. "reportdb/utils"
)

func DistributeQuery(queryChannel chan QueryReceive, readers []*Reader) {

	go func() {

		defer ShutdownReaders(readers)

		numReaders := uint8(len(readers))

		for query := range queryChannel {

			index := uint8((query.RequestID) % uint64(numReaders))

			if index >= numReaders || index < 0 {

				Logger.Warn("DistributeQuery : Query index is out of range",
					zap.Uint8("index", index),
					zap.Uint8("numReaders", numReaders),
					zap.Uint64("RequestID", query.RequestID),
				)

				continue
			}

			readers[index].queryEvents <- query
		}
	}()
}
