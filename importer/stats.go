package nebula_importer

import (
	"log"
	"time"
)

func InitStatsWorker(ch <-chan Stats) {
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		now := time.Now()
		var (
			totalCount, totalLatency uint64  = 0, 0
			totalReqTime             float64 = 0.0
		)
		for {
			select {
			case <-ticker.C:
				if totalCount == 0 {
					continue
				}
				secs := time.Since(now).Seconds()
				avgLatency := totalLatency / totalCount
				avgReq := 1000000 * totalReqTime / float64(totalCount)
				qps := float64(totalCount) / secs
				log.Printf("Requests: finished(%d), latency AVG(%dus), req AVG(%.2fus), QPS(%.2f/s)", totalCount, avgLatency, avgReq, qps)
			case stat := <-ch:
				if stat.Done {
					return
				}
				totalCount++
				totalReqTime += stat.ReqTime
				totalLatency += stat.Latency
			}
		}
	}()
}
