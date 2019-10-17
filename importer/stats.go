package nebula_importer

import (
	"log"
	"time"
)

func InitStatsWorker(ch <-chan Stats) {
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		var totalCount, totalLatency uint64 = 0, 0
		var totalReqTime float64 = 0.0
		for {
			select {
			case <-ticker.C:
				log.Printf("nebula requests: finished(%d), latency AVG(%dus), req AVG(%.2fus), QPS(%.2f/s)",
					totalCount, totalLatency/totalCount, totalReqTime*1000*1000/float64(totalCount), float64(totalCount)/totalReqTime)
			case stat := <-ch:
				totalCount++
				totalReqTime += stat.ReqTime
				totalLatency += stat.Latency
			}
		}
	}()
}
