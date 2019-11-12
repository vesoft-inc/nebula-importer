package stats

import (
	"time"

	"github.com/vesoft-inc/nebula-importer/pkg/base"
	"github.com/vesoft-inc/nebula-importer/pkg/logger"
)

type StatsMgr struct {
	StatsCh      chan base.Stats
	FileDoneCh   chan bool
	totalCount   uint64
	totalBatches uint64
	totalLatency uint64
	numFailed    uint64
	totalReqTime float64
}

func NewStatsMgr() *StatsMgr {
	m := StatsMgr{
		StatsCh:      make(chan base.Stats),
		FileDoneCh:   make(chan bool),
		totalCount:   0,
		totalLatency: 0,
		totalBatches: 0,
		numFailed:    0,
		totalReqTime: 0.0,
	}
	m.initStatsWorker()
	return &m
}

func (s *StatsMgr) Close() {
	close(s.StatsCh)
}

func (s *StatsMgr) updateStat(stat base.Stats) {
	s.totalBatches++
	s.totalCount += uint64(stat.BatchSize)
	s.totalReqTime += stat.ReqTime
	s.totalLatency += stat.Latency
}

func (s *StatsMgr) updateFailed(stat base.Stats) {
	s.totalBatches++
	s.totalCount += uint64(stat.BatchSize)
	s.numFailed += uint64(stat.BatchSize)
}

func (s *StatsMgr) print(now time.Time) {
	if s.totalCount == 0 {
		return
	}
	secs := time.Since(now).Seconds()
	avgLatency := s.totalLatency / s.totalBatches
	avgReq := 1000000 * s.totalReqTime / float64(s.totalBatches)
	qps := float64(s.totalCount) / secs
	logger.Log.Printf("Time(%.2fs), Finished(%d), Failed(%d), Latency AVG(%dus), Batches Req AVG(%.2fus), QPS(%.2f/s)",
		secs, s.totalCount, s.numFailed, avgLatency, avgReq, qps)
}

func (s *StatsMgr) initStatsWorker() {
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		now := time.Now()
		for {
			select {
			case <-ticker.C:
				s.print(now)
			case stat, ok := <-s.StatsCh:
				if !ok {
					return
				}
				switch stat.Type {
				case base.SUCCESS:
					s.updateStat(stat)
				case base.FAILURE:
					s.updateFailed(stat)
				case base.FILEDONE:
					s.print(now)
					s.FileDoneCh <- true
				default:
					logger.Log.Fatalf("Error stats type: %s", stat.Type)
				}
			}
		}
	}()
}
