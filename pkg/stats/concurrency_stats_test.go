package stats

import (
	"math/rand"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ConcurrencyStats", func() {
	It("concurrency", func() {
		rand.Seed(time.Now().UnixNano())

		concurrencyStats := NewConcurrencyStats()
		concurrencyStats.Init()
		initStats := concurrencyStats.Stats()
		Expect(initStats.StartTime.IsZero()).To(BeFalse())
		Expect(initStats.TotalRecords).To(BeZero())
		Expect(initStats.Percentage()).To(Equal(0.0))

		concurrency := 100
		totalBytes := make([]int64, concurrency)
		failedBytes := make([]int64, concurrency)
		succeededBytes := make([]int64, concurrency)
		var (
			sumBytes         int64
			sumFailedRecords int64
			sumRecords       int64
			sumFailedBatches int64
			sumBatches       int64
		)

		var wg sync.WaitGroup

		for i := 0; i < concurrency; i++ {
			totalBytes[i] = rand.Int63n(int64(1024*1024*1024)) + 1024
			sumBytes += totalBytes[i]

			wg.Add(1)
			go func(index int) {
				concurrencyStats.AddTotalBytes(totalBytes[index])
				wg.Done()
			}(i)

			if rand.Intn(2) > 0 {
				failedBytes[i] = rand.Int63n(totalBytes[i]) + 1

				sumFailedRecords += 7
				sumRecords += 7
				sumFailedBatches += 1
				sumBatches += 1

				wg.Add(1)
				go func(nBytes int64) {
					concurrencyStats.RequestFailed(7)
					concurrencyStats.RequestFailed(7)
					concurrencyStats.Failed(nBytes, 7)
					wg.Done()
				}(failedBytes[i])
			}

			succeededBytes[i] = totalBytes[i] - failedBytes[i]
			if succeededBytes[i] > 0 {
				sumRecords += 7
				sumBatches += 1

				wg.Add(1)
				go func(nBytes int64) {
					concurrencyStats.RequestSucceeded(7, 9*time.Millisecond, 11*time.Millisecond)
					concurrencyStats.RequestSucceeded(7, 9*time.Millisecond, 11*time.Millisecond)
					concurrencyStats.Succeeded(nBytes, 7)
					wg.Done()
				}(succeededBytes[i])
			}

			wg.Add(1)
			go func() {
				_ = concurrencyStats.Stats()
				_ = concurrencyStats.String()
				wg.Done()
			}()
		}

		wg.Wait()
		concurrencyStats.Init()
		s := concurrencyStats.Stats()
		Expect(s).To(Equal(&Stats{
			StartTime:       initStats.StartTime,
			ProcessedBytes:  sumBytes,
			TotalBytes:      sumBytes,
			FailedRecords:   sumFailedRecords,
			TotalRecords:    sumRecords,
			FailedRequest:   sumFailedBatches * 2,
			TotalRequest:    sumBatches * 2,
			TotalLatency:    9 * time.Millisecond * time.Duration(sumBatches-sumFailedBatches) * 2,
			TotalRespTime:   11 * time.Millisecond * time.Duration(sumBatches-sumFailedBatches) * 2,
			FailedProcessed: sumFailedRecords * 2,
			TotalProcessed:  sumRecords * 2,
		}))

		Expect(s.Percentage()).To(Equal(100.0))
		Expect(s.String()).To(ContainSubstring("100.00%("))
	})
})
