package writer

import (
	"go.uber.org/zap"
	. "reportdb/logger"
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

					Logger.Warn("DistributeData: writer index out of range",
						zap.Uint8("index", index),
						zap.Uint8("numWriters", numWriters),
						zap.Uint16("CounterId", row.CounterId),
						zap.Uint32("ObjectId", row.ObjectId),
					)

					continue
				}

				writers[index].events <- row
			}
		}

		return

	}()

	return
}
