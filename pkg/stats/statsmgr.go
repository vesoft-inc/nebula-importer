package stats

import (
	"fmt"
	"time"

	"github.com/vesoft-inc/nebula-importer/pkg/base"
	"github.com/vesoft-inc/nebula-importer/pkg/logger"
)

type StatsMgr struct {
	StatsCh         chan base.Stats
	DoneCh          chan bool
	totalCount      int64
	totalBatches    int64
	totalLatency    int64
	numFailed       int64
	totalReqTime    int64
	numReadingFiles int
}

func NewStatsMgr(numReadingFiles int) *StatsMgr {
	m := StatsMgr{
		StatsCh:         make(chan base.Stats),
		DoneCh:          make(chan bool),
		totalCount:      0,
		totalLatency:    0,
		totalBatches:    0,
		numFailed:       0,
		totalReqTime:    0.0,
		numReadingFiles: numReadingFiles,
	}
	go m.startWorker()
	return &m
}

func (s *StatsMgr) Close() {
	close(s.StatsCh)
	close(s.DoneCh)
}

func (s *StatsMgr) updateStat(stat base.Stats) {
	s.totalBatches++
	s.totalCount += int64(stat.BatchSize)
	s.totalReqTime += stat.ReqTime
	s.totalLatency += stat.Latency
}

func (s *StatsMgr) updateFailed(stat base.Stats) {
	s.totalBatches++
	s.totalCount += int64(stat.BatchSize)
	s.numFailed += int64(stat.BatchSize)
}

func (s *StatsMgr) print(prefix string, now time.Time) {
	if s.totalCount == 0 {
		return
	}
	secs := time.Since(now).Seconds()
	avgLatency := s.totalLatency / s.totalBatches
	avgReq := s.totalReqTime / s.totalBatches
	qps := float64(s.totalCount) / secs
	logger.Log.Printf("%s: Time(%.2fs), Finished(%d), Failed(%d), Latency AVG(%dus), Batches Req AVG(%dus), QPS(%.2f/s)",
		prefix, secs, s.totalCount, s.numFailed, avgLatency, avgReq, qps)
}

func (s *StatsMgr) startWorker() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	now := time.Now()
	for {
		select {
		case <-ticker.C:
			s.print("Tick", now)
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
				s.print(fmt.Sprintf("Done(%s)", stat.Filename), now)
				s.numReadingFiles--
				if s.numReadingFiles == 0 {
					s.DoneCh <- true
				}
			default:
				logger.Log.Fatalf("Error stats type: %s", stat.Type)
			}
		}
	}
}
