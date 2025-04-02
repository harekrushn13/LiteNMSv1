package query

import (
	"fmt"
	"reportdb/config"
	"reportdb/src/datastore/reader"
	"sync"
	"time"
)

func RunQuery(wg *sync.WaitGroup) {

	defer wg.Done()

	t := time.NewTicker(config.PollingInterval)

	defer t.Stop()

	to := time.Now().Unix()

	stopTimer := time.NewTimer(config.StopTime)

	defer stopTimer.Stop()

	for {
		select {

		case <-t.C:
			v, err := reader.FetchData(3, 3, 1743569072, uint32(to)+5)

			if err != nil {

				fmt.Println(err)
			}

			fmt.Printf("%#v\n", v)

		case <-stopTimer.C:

			return
		}
	}

}
