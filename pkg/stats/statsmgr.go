package stats

import (
	"fmt"
	"time"

	"github.com/vesoft-inc/nebula-importer/pkg/base"
	"github.com/vesoft-inc/nebula-importer/pkg/config"
	"github.com/vesoft-inc/nebula-importer/pkg/logger"
	"github.com/vesoft-inc/nebula-importer/pkg/reader"
)

type StatsMgr struct {
	OutputStatsCh chan Stats
	StatsCh       chan base.Stats
	DoneCh        chan bool
	Stats         Stats
	Done          bool
	CountFileDone bool
}

type Stats struct {
	NumFailed          int64 `json:"numFailed"`
	NumReadFailed      int64 `json:"numReadFailed"`
	TotalCount         int64 `json:"totalCount"`
	TotalBatches       int64 `json:"totalBatches"`
	TotalLatency       int64 `json:"totalLatency"`
	TotalReqTime       int64 `json:"totalReqTime"`
	TotalBytes         int64 `json:"totalBytes"`
	TotalImportedBytes int64 `json:"totalImportedBytes"`
}

func NewStatsMgr(files []*config.File) *StatsMgr {
	numReadingFiles := len(files)
	stats := Stats{
		NumFailed:    0,
		TotalBytes:   0,
		TotalCount:   0,
		TotalLatency: 0,
		TotalBatches: 0,
		TotalReqTime: 0.0,
	}
	m := StatsMgr{
		OutputStatsCh: make(chan Stats),
		StatsCh:       make(chan base.Stats),
		DoneCh:        make(chan bool),
		Stats:         stats,
	}
	go m.startWorker(numReadingFiles)
	return &m
}

func (s *StatsMgr) Close() {
	close(s.StatsCh)
	close(s.DoneCh)
	close(s.OutputStatsCh)
	s.Done = true
}

func (s *StatsMgr) updateStat(stat base.Stats) {
	s.Stats.TotalBatches++
	s.Stats.TotalCount += int64(stat.BatchSize)
	s.Stats.TotalReqTime += stat.ReqTime
	s.Stats.TotalLatency += stat.Latency
	s.Stats.TotalImportedBytes += stat.ImportedBytes
}

func (s *StatsMgr) updateFailed(stat base.Stats) {
	s.Stats.TotalBatches++
	s.Stats.TotalCount += int64(stat.BatchSize)
	s.Stats.NumFailed += int64(stat.BatchSize)
	s.Stats.TotalImportedBytes += stat.ImportedBytes
}

func (s *StatsMgr) outputStats() {
	s.OutputStatsCh <- s.Stats
}

func (s *StatsMgr) print(prefix string, now time.Time) {
	if s.Stats.TotalCount == 0 {
		return
	}
	secs := time.Since(now).Seconds()
	avgLatency := s.Stats.TotalLatency / s.Stats.TotalBatches
	avgReq := s.Stats.TotalReqTime / s.Stats.TotalBatches
	rps := float64(s.Stats.TotalCount) / secs
	logger.Infof("%s: Time(%.2fs), Finished(%d), Failed(%d), Read Failed(%d), Latency AVG(%dus), Batches Req AVG(%dus), Rows AVG(%.2f/s)",
		prefix, secs, s.Stats.TotalCount, s.Stats.NumFailed, s.Stats.NumReadFailed, avgLatency, avgReq, rps)
}

func (s *StatsMgr) CountFileBytes(freaders []*reader.FileReader) error {
	if s.CountFileDone {
		return nil
	}
	s.Stats.TotalBytes = 0
	for _, r := range freaders {
		if r == nil {
			continue
		}
		bytes, err := r.DataReader.TotalBytes()
		if err != nil {
			return err
		}
		s.Stats.TotalBytes += bytes
	}
	s.CountFileDone = true
	return nil
}

func (s *StatsMgr) startWorker(numReadingFiles int) {
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
				numReadingFiles--
				if numReadingFiles == 0 {
					s.DoneCh <- true
				}
			case base.OUTPUT:
				s.outputStats()
			default:
				logger.Errorf("Error stats type: %s", stat.Type)
			}
		}
	}
}
