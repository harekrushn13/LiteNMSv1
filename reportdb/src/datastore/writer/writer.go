package writer

import (
	"log"
	"reportdb/config"
	"reportdb/src/storage/engine"
	"sync"
)

func StartWriter(pollCh <-chan []config.RowData, wg *sync.WaitGroup) {

	writerChannels := make([]chan config.RowData, config.WriterCount)

	for i := uint8(0); i < config.WriterCount; i++ {

		writerChannels[i] = make(chan config.RowData, 100)

		wg.Add(1)

		go runWriter(writerChannels[i], wg)
	}

	go func() {

		for batch := range pollCh {

			for _, row := range batch {

				writerIdx := uint8((uint32(row.CounterId) + row.ObjectId) % uint32(config.WriterCount))

				writerChannels[writerIdx] <- row
			}
		}

		for _, ch := range writerChannels {

			close(ch)
		}
	}()
}

func runWriter(ch <-chan config.RowData, wg *sync.WaitGroup) {

	defer wg.Done()

	for row := range ch {

		if err := writeRow(row); err != nil {

			log.Printf("Error writing row: %v", err)
		}
	}
}

func writeRow(row config.RowData) error {

	err := engine.Save(row)

	if err != nil {

		return err
	}

	return nil
}
