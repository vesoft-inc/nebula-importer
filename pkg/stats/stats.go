package stats

import (
	"fmt"
	"time"
)

type Stats struct {
	Latency uint64
	ReqTime float64
}

type count struct {
	totalCount   uint64
	totalLatency uint64
	numFailed    uint64
	totalReqTime float64
}

func newCount() count {
	return count{
		totalCount:   0,
		totalLatency: 0,
		numFailed:    0,
		totalReqTime: 0.0,
	}
}

func (s *count) updateStat(stat stats.Stats) {
	s.totalCount++
	s.totalReqTime += stat.ReqTime
	s.totalLatency += stat.Latency
}

func (s *count) updateFailed() {
	s.totalCount++
	s.numFailed++
}

func (s *count) print(now time.Time) {
	if s.totalCount == 0 {
		return
	}
	secs := time.Since(now).Seconds()
	avgLatency := s.totalLatency / s.totalCount
	avgReq := 1000000 * s.totalReqTime / float64(s.totalCount)
	qps := float64(s.totalCount) / secs
	fmt.Printf("\rRequests: finished(%d), Failed(%d), latency AVG(%dus), req AVG(%.2fus), QPS(%.2f/s)",
		s.totalCount, s.numFailed, avgLatency, avgReq, qps)
}

func InitStatsWorker(ch <-chan stats.Stats, failCh <-chan bool) {
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		c := newCount()
		now := time.Now()
		for {
			select {
			case <-ticker.C:
				c.print(now)
			case stat, ok := <-ch:
				if !ok {
					c.print(now)
					return
				}
				c.updateStat(stat)
			case <-failCh:
				c.updateFailed()
			}
		}
	}()
}
