package stats

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Stats", func() {
	Describe(".String", func() {
		It("TotalRecords is zero", func() {
			s := &Stats{
				StartTime: time.Now(),
			}
			Expect(s.String()).Should(Equal("0s ... 0.00%(0 B/0 B) Records{Finished: 0, Failed: 0, Rate: 0.00/s}, Requests{Finished: 0, Failed: 0, Latency: 0s/0s, Rate: 0.00/s}, Processed{Finished: 0, Failed: 0, Rate: 0.00/s}"))
		})
		It("TotalRecords is not zero", func() {
			s := &Stats{
				StartTime:       time.Now().Add(-time.Second * 10),
				ProcessedBytes:  100 * 1024,
				TotalBytes:      300 * 1024,
				FailedRecords:   23,
				TotalRecords:    1234,
				FailedRequest:   1,
				TotalRequest:    12,
				TotalLatency:    time.Second * 12,
				TotalRespTime:   2 * time.Second * 12,
				FailedProcessed: 2,
				TotalProcessed:  5,
			}
			Expect(s.String()).Should(Equal("10s 20s 33.33%(100 KiB/300 KiB) Records{Finished: 1234, Failed: 23, Rate: 123.40/s}, Requests{Finished: 12, Failed: 1, Latency: 1s/2s, Rate: 1.20/s}, Processed{Finished: 5, Failed: 2, Rate: 0.50/s}"))
		})
	})
})
