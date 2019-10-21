package importer

import (
	"fmt"
	"time"
)

func InitStatsWorker(ch <-chan Stats, failCh <-chan bool) {
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		now := time.Now()
		var (
			totalCount, totalLatency, numFailed uint64  = 0, 0, 0
			totalReqTime                        float64 = 0.0
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
				fmt.Printf("\rRequests: finished(%d), Failed(%d), latency AVG(%dus), req AVG(%.2fus), QPS(%.2f/s)", totalCount, numFailed, avgLatency, avgReq, qps)
			case stat, ok := <-ch:
				if !ok {
					return
				}
				totalCount++
				totalReqTime += stat.ReqTime
				totalLatency += stat.Latency
			case <-failCh:
				numFailed++
			}
		}
	}()
}
