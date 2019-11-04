package stats

import (
	"fmt"
	"log"
	"time"
)

type StatType int

const (
	SUCCESS StatType = 0
	FAILURE StatType = 1
	PRINT   StatType = 2
)

type Stats struct {
	Type      StatType
	Latency   uint64
	ReqTime   float64
	BatchSize int
}

func NewSuccessStats(latency uint64, reqTime float64) Stats {
	return Stats{
		Type:    SUCCESS,
		Latency: latency,
		ReqTime: reqTime,
	}
}

func NewFailureStats(batchSize int) Stats {
	return Stats{
		Type:      FAILURE,
		BatchSize: batchSize,
	}
}

type StatsMgr struct {
	statsCh      chan Stats
	totalCount   uint64
	totalLatency uint64
	numFailed    uint64
	totalReqTime float64
}

func NewStatsMgr() *StatsMgr {
	m := &StatsMgr{
		statsCh:      make(chan Stats),
		totalCount:   0,
		totalLatency: 0,
		numFailed:    0,
		totalReqTime: 0.0,
	}
	m.initStatsWorker()
	return m
}

func (s *StatsMgr) GetStatsChan() chan<- Stats {
	return s.statsCh
}

func (s *StatsMgr) Close() {
	close(s.statsCh)
}

func (s *StatsMgr) updateStat(stat Stats) {
	s.totalCount += uint64(stat.BatchSize)
	s.totalReqTime += stat.ReqTime
	s.totalLatency += stat.Latency
}

func (s *StatsMgr) updateFailed(stat Stats) {
	s.totalCount += uint64(stat.BatchSize)
	s.numFailed += uint64(stat.BatchSize)
}

func (s *StatsMgr) print(now time.Time) {
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

func (s *StatsMgr) initStatsWorker() {
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		now := time.Now()
		for {
			select {
			case <-ticker.C:
				s.print(now)
			case stat, ok := <-s.statsCh:
				if !ok {
					s.print(now)
					return
				}
				switch stat.Type {
				case SUCCESS:
					s.updateStat(stat)
				case FAILURE:
					s.updateFailed(stat)
				case PRINT:
					s.print(now)
				default:
					log.Fatalf("Error stats type: %s", stat.Type)
				}
			}
		}
	}()
}
