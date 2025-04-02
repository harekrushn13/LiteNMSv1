package main

import (
	"log"
	"reportdb/profiling"
	"reportdb/src/datastore/query"
	"reportdb/src/datastore/writer"
	"reportdb/src/polling"
	"reportdb/src/utils"
	"sync"
	"time"
)

func main() {
	cpuProfileFile := profiling.StartCPUProfile()
	//
	defer profiling.StopCPUProfile(cpuProfileFile)

	//defer profiling.WriteMemProfile()

	//defer profiling.WriteGoroutineProfile()

	//traceFile := profiling.StartTrace()
	//
	//defer profiling.StopTrace(traceFile)

	pollCh := polling.PollData()

	var wg sync.WaitGroup

	writer.StartWriter(pollCh, &wg)

	// save Index at specific Interval

	wg.Add(1)

	go func() {

		defer wg.Done()

		t := time.NewTicker(2 * time.Second)

		defer t.Stop()

		stopTimer := time.NewTimer(12 * time.Second)

		defer stopTimer.Stop()

		for {

			select {

			case <-t.C:

				if err := utils.SaveIndex(time.Now()); err != nil {

					log.Fatal(err)
				}

			case <-stopTimer.C:

				if err := utils.SaveIndex(time.Now()); err != nil {

					log.Fatal(err)
				}

				return
			}
		}
	}()

	// parallel reader

	func() {

		for i := 0; i < 3; i++ {

			wg.Add(1)

			go query.RunQuery(&wg)
		}

	}()

	wg.Wait()
}
