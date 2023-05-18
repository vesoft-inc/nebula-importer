package stats

import (
	"sync"
	"time"
)

type (
	ConcurrencyStats struct {
		s       Stats
		mu      sync.Mutex
		initOne sync.Once
	}
)

func NewConcurrencyStats() *ConcurrencyStats {
	return &ConcurrencyStats{}
}

func (s *ConcurrencyStats) Init() {
	s.initOne.Do(func() {
		s.s.StartTime = time.Now()
	})
}

func (s *ConcurrencyStats) AddTotalBytes(nBytes int64) {
	s.mu.Lock()
	s.s.TotalBytes += nBytes
	s.mu.Unlock()
}

func (s *ConcurrencyStats) Failed(nBytes, nRecords int64) {
	s.mu.Lock()
	s.s.ProcessedBytes += nBytes
	s.s.FailedRecords += nRecords
	s.s.TotalRecords += nRecords
	s.mu.Unlock()
}

func (s *ConcurrencyStats) Succeeded(nBytes, nRecords int64) {
	s.mu.Lock()
	s.s.ProcessedBytes += nBytes
	s.s.TotalRecords += nRecords
	s.mu.Unlock()
}

func (s *ConcurrencyStats) RequestFailed(nRecords int64) {
	s.mu.Lock()
	s.s.FailedRequest++
	s.s.TotalRequest++
	s.s.FailedProcessed += nRecords
	s.s.TotalProcessed += nRecords
	s.mu.Unlock()
}

func (s *ConcurrencyStats) RequestSucceeded(nRecords int64, latency, respTime time.Duration) {
	s.mu.Lock()
	s.s.TotalRequest++
	s.s.TotalLatency += latency
	s.s.TotalRespTime += respTime
	s.s.TotalProcessed += nRecords
	s.mu.Unlock()
}

func (s *ConcurrencyStats) Stats() *Stats {
	s.mu.Lock()
	cpy := s.s
	s.mu.Unlock()
	return &cpy
}

func (s *ConcurrencyStats) String() string {
	return s.Stats().String()
}
