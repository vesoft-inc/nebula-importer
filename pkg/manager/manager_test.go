package manager

import (
	stderrors "errors"
	"io"
	"os"
	"sync/atomic"
	"time"

	"github.com/vesoft-inc/nebula-importer/v4/pkg/client"
	"github.com/vesoft-inc/nebula-importer/v4/pkg/importer"
	"github.com/vesoft-inc/nebula-importer/v4/pkg/logger"
	"github.com/vesoft-inc/nebula-importer/v4/pkg/reader"
	"github.com/vesoft-inc/nebula-importer/v4/pkg/source"
	"github.com/vesoft-inc/nebula-importer/v4/pkg/spec"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Manager", func() {
	It("New", func() {
		m := New(client.NewPool())
		m1, ok := m.(*defaultManager)
		Expect(ok).To(BeTrue())
		Expect(m1).NotTo(BeNil())
		Expect(m1.pool).NotTo(BeNil())
		Expect(m1.getClientOptions).To(BeNil())
		Expect(m1.batch).To(Equal(0))
		Expect(m1.readerConcurrency).To(Equal(DefaultReaderConcurrency))
		Expect(m1.readerPool).NotTo(BeNil())
		Expect(m1.importerConcurrency).To(Equal(DefaultImporterConcurrency))
		Expect(m1.importerPool).NotTo(BeNil())
		Expect(m1.statsInterval).To(Equal(DefaultStatsInterval))
		Expect(m1.hooks.Before).To(BeEmpty())
		Expect(m1.hooks.After).To(BeEmpty())
		Expect(m1.logger).NotTo(BeNil())
	})

	It("NewWithOpts", func() {
		m := NewWithOpts(
			WithGraphName("graphName"),
			WithClientPool(client.NewPool()),
			WithGetClientOptions(client.WithClientInitFunc(nil)),
			WithBatch(1),
			WithReaderConcurrency(DefaultReaderConcurrency+1),
			WithImporterConcurrency(DefaultImporterConcurrency+1),
			WithStatsInterval(DefaultStatsInterval+1),
			WithBeforeHooks(&Hook{
				Statements: []string{"before statements1"},
				Wait:       time.Second,
			}),
			WithAfterHooks(&Hook{
				Statements: []string{"after statements"},
				Wait:       time.Second,
			}),
			WithLogger(logger.NopLogger),
		)
		m1, ok := m.(*defaultManager)
		Expect(ok).To(BeTrue())
		Expect(m1).NotTo(BeNil())
		Expect(m1.pool).NotTo(BeNil())
		Expect(m1.getClientOptions).NotTo(BeNil())
		Expect(m1.batch).To(Equal(1))
		Expect(m1.readerConcurrency).To(Equal(DefaultReaderConcurrency + 1))
		Expect(m1.readerPool).NotTo(BeNil())
		Expect(m1.importerConcurrency).To(Equal(DefaultImporterConcurrency + 1))
		Expect(m1.importerPool).NotTo(BeNil())
		Expect(m1.statsInterval).To(Equal(DefaultStatsInterval + 1))
		Expect(m1.hooks.Before).To(HaveLen(1))
		Expect(m1.hooks.After).To(HaveLen(1))
		Expect(m1.logger).NotTo(BeNil())
	})

	Describe("Run", func() {
		var (
			tmpdir                string
			ctrl                  *gomock.Controller
			mockSource            *source.MockSource
			mockBatchRecordReader *reader.MockBatchRecordReader
			mockClient            *client.MockClient
			mockClientPool        *client.MockPool
			mockResponse          *client.MockResponse
			mockImporter          *importer.MockImporter
			m                     Manager
			batch                 = 10
		)
		BeforeEach(func() {
			var err error
			tmpdir, err = os.MkdirTemp("", "test")
			Expect(err).NotTo(HaveOccurred())

			ctrl = gomock.NewController(GinkgoT())
			mockSource = source.NewMockSource(ctrl)
			mockBatchRecordReader = reader.NewMockBatchRecordReader(ctrl)
			mockClient = client.NewMockClient(ctrl)
			mockClientPool = client.NewMockPool(ctrl)
			mockResponse = client.NewMockResponse(ctrl)
			mockImporter = importer.NewMockImporter(ctrl)

			l, err := logger.New(logger.WithLevel(logger.WarnLevel))
			Expect(err).NotTo(HaveOccurred())
			m = New(
				mockClientPool,
				WithBatch(batch),
				WithLogger(l),
				WithBeforeHooks(&Hook{
					Statements: []string{"before statement"},
					Wait:       time.Second,
				}),
				WithAfterHooks(&Hook{
					Statements: []string{"after statement"},
					Wait:       time.Second,
				}),
			)
		})

		AfterEach(func() {
			ctrl.Finish()
			err := os.RemoveAll(tmpdir)
			Expect(err).NotTo(HaveOccurred())
		})

		It("concurrency successfully", func() {
			var err error
			loopCountPreFile := 10

			fnNewSource := func() source.Source {
				mockSource = source.NewMockSource(ctrl)

				mockSource.EXPECT().Name().Times(2).Return("source name")
				mockSource.EXPECT().Open().Times(1).Return(nil)
				mockSource.EXPECT().Size().Times(1).Return(int64(12345), nil)
				mockSource.EXPECT().Close().Times(1).Return(nil)

				return mockSource
			}

			fnNewBatchRecordReader := func(count int64) reader.BatchRecordReader {
				mockBatchRecordReader = reader.NewMockBatchRecordReader(ctrl)
				var currBatchRecordReaderCount int64
				fnReadBatch := func() (int, spec.Records, error) {
					if curr := atomic.AddInt64(&currBatchRecordReaderCount, 1); curr > count {
						return 0, nil, io.EOF
					}
					return 11, spec.Records{
						[]string{"0123"},
						[]string{"4567"},
						[]string{"890"},
					}, nil
				}
				mockBatchRecordReader.EXPECT().ReadBatch().Times(int(count) + 1).DoAndReturn(fnReadBatch)
				return mockBatchRecordReader
			}

			var (
				batchRecordReaderNodeCount int64 = 1017
				batchRecordReaderEdgeCount int64 = 1037
				executeFailedTimes         int64 = 10
				currExecuteTimes           int64
			)

			totalBatches := (batchRecordReaderNodeCount + batchRecordReaderEdgeCount) * int64(loopCountPreFile)
			processedBytes := (11*batchRecordReaderNodeCount + 11*batchRecordReaderEdgeCount) * int64(loopCountPreFile)
			totalBytes := 12345 * int64(loopCountPreFile) * 2
			totalRecords := totalBatches * 3

			fnImport := func(records ...spec.Record) (*importer.ImportResp, error) {
				curr := atomic.AddInt64(&currExecuteTimes, 1)
				if curr%100 == 0 && curr/100 <= executeFailedTimes {
					return nil, stderrors.New("import failed")
				}
				return &importer.ImportResp{
					Latency:  2 * time.Microsecond,
					RespTime: 3 * time.Microsecond,
				}, nil
			}

			gomock.InOrder(
				mockClientPool.EXPECT().GetClient(gomock.Any()).Return(mockClient, nil),
				mockClient.EXPECT().Execute("before statement").Times(1).Return(mockResponse, nil),
				mockResponse.EXPECT().IsSucceed().Return(true),

				mockClientPool.EXPECT().Open().Return(nil),

				mockClientPool.EXPECT().GetClient(gomock.Any()).Return(mockClient, nil),
				mockClient.EXPECT().Execute("after statement").Times(1).Return(mockResponse, nil),
				mockResponse.EXPECT().IsSucceed().Return(true),
			)

			mockImporter.EXPECT().Import(gomock.Any()).AnyTimes().DoAndReturn(fnImport)
			mockImporter.EXPECT().Add(1).Times(loopCountPreFile*4 + int(totalBatches)*2)
			mockImporter.EXPECT().Done().Times(loopCountPreFile*4 + int(totalBatches)*2)
			mockImporter.EXPECT().Wait().Times(loopCountPreFile * 4)

			for i := 0; i < loopCountPreFile; i++ {
				err = m.Import(
					fnNewSource(),
					fnNewBatchRecordReader(batchRecordReaderNodeCount),
					mockImporter,
					mockImporter,
				)
				Expect(err).NotTo(HaveOccurred())
				err = m.Import(
					fnNewSource(),
					fnNewBatchRecordReader(batchRecordReaderEdgeCount),
					mockImporter,
					mockImporter,
				)
				Expect(err).NotTo(HaveOccurred())
			}

			err = m.Start()
			Expect(err).NotTo(HaveOccurred())

			err = m.Wait()
			Expect(err).NotTo(HaveOccurred())
			s := m.Stats()

			Expect(s.StartTime.IsZero()).To(BeFalse())
			Expect(s.ProcessedBytes).To(Equal(processedBytes))
			Expect(s.TotalBytes).To(Equal(totalBytes))
			Expect(s.FailedRecords).NotTo(Equal(int64(0)))
			Expect(s.FailedRecords).To(BeNumerically("<=", executeFailedTimes*int64(batch)))
			Expect(s.TotalRecords).To(Equal(totalRecords))
			Expect(s.FailedRequest).To(Equal(executeFailedTimes))
			Expect(s.TotalRequest).To(Equal(totalBatches * 2))
			Expect(s.TotalLatency).To(Equal(2 * time.Microsecond * time.Duration((totalBatches*2)-executeFailedTimes)))
			Expect(s.TotalRespTime).To(Equal(3 * time.Microsecond * time.Duration((totalBatches*2)-executeFailedTimes)))
			Expect(s.FailedProcessed).NotTo(Equal(int64(0)))
			Expect(s.FailedRecords).To(BeNumerically("<=", executeFailedTimes*int64(batch)))
			Expect(s.TotalProcessed).To(Equal(totalRecords * 2))
		})

		It("Import without importer", func() {
			err := m.Import(
				mockSource,
				mockBatchRecordReader,
			)
			Expect(err).NotTo(HaveOccurred())
		})

		It("source open failed", func() {
			mockSource.EXPECT().Name().Return("source name")
			mockSource.EXPECT().Open().Return(os.ErrNotExist)

			err := m.Import(
				mockSource,
				mockBatchRecordReader,
				mockImporter,
			)
			Expect(err).To(HaveOccurred())
			Expect(stderrors.Is(err, os.ErrNotExist)).To(BeTrue())
		})

		It("get client failed", func() {
			mockClientPool.EXPECT().GetClient(gomock.Any()).Return(nil, stderrors.New("test error"))

			err := m.Start()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("test error"))
		})

		It("exec before failed", func() {
			gomock.InOrder(
				mockClientPool.EXPECT().GetClient(gomock.Any()).Return(mockClient, nil),
				mockClient.EXPECT().Execute("before statement").Times(1).Return(nil, stderrors.New("test error")),
			)

			err := m.Start()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("test error"))
		})

		It("client pool open failed", func() {
			gomock.InOrder(
				mockClientPool.EXPECT().GetClient(gomock.Any()).Return(mockClient, nil),
				mockClient.EXPECT().Execute("before statement").Times(1).Return(mockResponse, nil),
				mockResponse.EXPECT().IsSucceed().Return(true),

				mockClientPool.EXPECT().Open().Return(stderrors.New("test error")),
			)

			err := m.Start()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("test error"))
		})

		It("exec after failed", func() {
			gomock.InOrder(
				mockClientPool.EXPECT().GetClient(gomock.Any()).Return(mockClient, nil),
				mockClient.EXPECT().Execute("before statement").Times(1).Return(mockResponse, nil),
				mockResponse.EXPECT().IsSucceed().Return(true),

				mockClientPool.EXPECT().Open().Return(nil),

				mockClientPool.EXPECT().GetClient(gomock.Any()).Return(mockClient, nil),
				mockClient.EXPECT().Execute("after statement").Times(1).Return(nil, stderrors.New("test error")),
			)

			err := m.Start()
			Expect(err).NotTo(HaveOccurred())

			err = m.Wait()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("test error"))
		})

		It("stop successfully", func() {
			gomock.InOrder(
				mockClientPool.EXPECT().GetClient(gomock.Any()).Return(mockClient, nil),
				mockClient.EXPECT().Execute("before statement").Times(1).Return(mockResponse, nil),
				mockResponse.EXPECT().IsSucceed().Return(true),

				mockClientPool.EXPECT().Open().Return(nil),

				mockClientPool.EXPECT().GetClient(gomock.Any()).Return(mockClient, nil),
				mockClient.EXPECT().Execute("after statement").Times(1).Return(mockResponse, nil),
				mockResponse.EXPECT().IsSucceed().Return(true),
			)

			err := m.Start()
			Expect(err).NotTo(HaveOccurred())

			time.Sleep(100 * time.Millisecond)

			err = m.Stop()
			Expect(err).NotTo(HaveOccurred())
		})

		It("stop failed", func() {
			gomock.InOrder(
				mockClientPool.EXPECT().GetClient(gomock.Any()).Return(mockClient, nil),
				mockClient.EXPECT().Execute("before statement").Times(1).Return(mockResponse, nil),
				mockResponse.EXPECT().IsSucceed().Return(true),

				mockClientPool.EXPECT().Open().Return(nil),

				mockClientPool.EXPECT().GetClient(gomock.Any()).Return(mockClient, nil),
				mockClient.EXPECT().Execute("after statement").Times(1).Return(mockResponse, nil),
				mockResponse.EXPECT().IsSucceed().Return(false),
				mockResponse.EXPECT().GetError().Times(1).Return(stderrors.New("exec failed")),
			)

			err := m.Start()
			Expect(err).NotTo(HaveOccurred())

			time.Sleep(100 * time.Millisecond)

			err = m.Stop()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("exec failed"))
			_ = m.Stop()
		})

		It("stop without read finished", func() {
			gomock.InOrder(
				mockClientPool.EXPECT().GetClient(gomock.Any()).Return(mockClient, nil),
				mockClient.EXPECT().Execute("before statement").Times(1).Return(mockResponse, nil),
				mockResponse.EXPECT().IsSucceed().Return(true),

				mockClientPool.EXPECT().Open().Return(nil),

				mockClientPool.EXPECT().GetClient(gomock.Any()).Return(mockClient, nil),
				mockClient.EXPECT().Execute("after statement").Times(1).Return(mockResponse, nil),
				mockResponse.EXPECT().IsSucceed().Return(true),
			)

			mockSource.EXPECT().Name().Times(2).Return("source name")
			mockSource.EXPECT().Open().Return(nil)
			mockSource.EXPECT().Size().Return(int64(1024*1024*1024*1024), nil)
			mockSource.EXPECT().Close().Return(nil)

			mockBatchRecordReader.EXPECT().ReadBatch().AnyTimes().Return(11, spec.Records{
				[]string{"0123"},
				[]string{"4567"},
				[]string{"890"},
			}, nil)

			mockImporter.EXPECT().Import(gomock.Any()).AnyTimes().Return(&importer.ImportResp{}, nil)
			mockImporter.EXPECT().Add(1).MinTimes(1)
			mockImporter.EXPECT().Done().MinTimes(1)
			mockImporter.EXPECT().Wait().Times(1)

			err := m.Import(
				mockSource,
				mockBatchRecordReader,
				mockImporter,
			)
			Expect(err).NotTo(HaveOccurred())

			err = m.Start()
			Expect(err).NotTo(HaveOccurred())

			time.Sleep(100 * time.Millisecond)

			err = m.Stop()
			Expect(err).NotTo(HaveOccurred())
		})

		It("no hooks", func() {
			m.(*defaultManager).hooks.Before = nil
			m.(*defaultManager).hooks.After = nil

			mockClientPool.EXPECT().Open().Return(nil)

			err := m.Start()
			Expect(err).NotTo(HaveOccurred())

			err = m.Wait()
			Expect(err).NotTo(HaveOccurred())
		})

		It("nil or empty hooks", func() {
			m.(*defaultManager).hooks.Before = []*Hook{
				nil,
				{Statements: []string{""}},
			}
			m.(*defaultManager).hooks.After = []*Hook{
				{Statements: []string{""}},
				nil,
			}

			mockClientPool.EXPECT().Open().Return(nil)

			err := m.Start()
			Expect(err).NotTo(HaveOccurred())

			err = m.Wait()
			Expect(err).NotTo(HaveOccurred())
		})

		It("disable stats interval", func() {
			m.(*defaultManager).hooks.Before = nil
			m.(*defaultManager).hooks.After = nil
			m.(*defaultManager).statsInterval = 0

			mockClientPool.EXPECT().Open().Return(nil)

			err := m.Start()
			Expect(err).NotTo(HaveOccurred())

			err = m.Wait()
			Expect(err).NotTo(HaveOccurred())
		})

		It("stats interval print", func() {
			m.(*defaultManager).hooks.Before = nil
			m.(*defaultManager).hooks.After = nil
			m.(*defaultManager).statsInterval = 10 * time.Microsecond

			mockClientPool.EXPECT().Open().Return(nil)

			err := m.Start()
			Expect(err).NotTo(HaveOccurred())

			time.Sleep(100 * time.Millisecond)

			err = m.Wait()
			Expect(err).NotTo(HaveOccurred())
		})

		It("submit reader failed", func() {
			m.(*defaultManager).hooks.Before = nil
			m.(*defaultManager).hooks.After = nil
			m.(*defaultManager).readerPool.Release()

			mockSource.EXPECT().Name().Times(2).Return("source name")
			mockSource.EXPECT().Open().Times(2).Return(nil)
			mockSource.EXPECT().Size().Times(2).Return(int64(1024), nil)
			mockSource.EXPECT().Close().Times(2).Return(nil)

			mockClientPool.EXPECT().Open().Return(nil)

			mockImporter.EXPECT().Add(1).Times(2)
			mockImporter.EXPECT().Done().Times(2)

			err := m.Import(
				mockSource,
				mockBatchRecordReader,
				mockImporter,
			)
			Expect(err).NotTo(HaveOccurred())
			err = m.Import(
				mockSource,
				mockBatchRecordReader,
				mockImporter,
			)
			Expect(err).NotTo(HaveOccurred())

			err = m.Start()
			Expect(err).NotTo(HaveOccurred())

			err = m.Wait()
			Expect(err).NotTo(HaveOccurred())
		})

		It("submit importer failed", func() {
			m.(*defaultManager).hooks.Before = nil
			m.(*defaultManager).hooks.After = nil
			m.(*defaultManager).importerPool.Release()

			mockSource.EXPECT().Name().Times(2 + 2).Return("source name")
			mockSource.EXPECT().Open().Times(2).Return(nil)
			mockSource.EXPECT().Size().Times(2).Return(int64(1024), nil)
			mockSource.EXPECT().Close().Times(2).Return(nil)

			gomock.InOrder(
				mockBatchRecordReader.EXPECT().ReadBatch().Times(2).Return(11, spec.Records{
					[]string{"0123"},
					[]string{"4567"},
					[]string{"890"},
				}, nil),
				mockBatchRecordReader.EXPECT().ReadBatch().Times(2).Return(0, spec.Records(nil), io.EOF),
			)

			mockClientPool.EXPECT().Open().Return(nil)

			mockImporter.EXPECT().Add(1).Times(2 + 2)
			mockImporter.EXPECT().Done().Times(2 + 2)
			mockImporter.EXPECT().Wait().Times(2)

			err := m.Import(
				mockSource,
				mockBatchRecordReader,
				mockImporter,
			)
			Expect(err).NotTo(HaveOccurred())
			err = m.Import(
				mockSource,
				mockBatchRecordReader,
				mockImporter,
			)
			Expect(err).NotTo(HaveOccurred())

			err = m.Start()
			Expect(err).NotTo(HaveOccurred())

			err = m.Wait()
			Expect(err).NotTo(HaveOccurred())
		})

		It("get size failed", func() {
			m.(*defaultManager).hooks.Before = nil
			m.(*defaultManager).hooks.After = nil

			mockSource.EXPECT().Name().Times(2).Return("source name")
			mockSource.EXPECT().Open().Times(2).Return(nil)
			mockSource.EXPECT().Size().Times(2).Return(int64(0), stderrors.New("test error"))
			mockSource.EXPECT().Close().Times(2).Return(nil)

			err := m.Import(
				mockSource,
				mockBatchRecordReader,
				mockImporter,
			)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("test error"))

			err = m.Import(
				mockSource,
				mockBatchRecordReader,
				mockImporter,
			)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("test error"))
		})

		It("read failed", func() {
			m.(*defaultManager).hooks.Before = nil
			m.(*defaultManager).hooks.After = nil

			mockClientPool.EXPECT().Open().Return(nil)
			mockSource.EXPECT().Name().Times(2 + 2).Return("source name")
			mockSource.EXPECT().Open().Times(2).Return(nil)
			mockSource.EXPECT().Size().Times(2).Return(int64(1024), nil)
			mockSource.EXPECT().Close().Times(2).Return(nil)

			mockBatchRecordReader.EXPECT().ReadBatch().Times(2).Return(0, spec.Records(nil), stderrors.New("test error"))

			mockImporter.EXPECT().Add(1).Times(2)
			mockImporter.EXPECT().Done().Times(2)
			mockImporter.EXPECT().Wait().Times(2)

			err := m.Import(
				mockSource,
				mockBatchRecordReader,
				mockImporter,
			)
			Expect(err).NotTo(HaveOccurred())
			err = m.Import(
				mockSource,
				mockBatchRecordReader,
				mockImporter,
			)
			Expect(err).NotTo(HaveOccurred())

			err = m.Start()
			Expect(err).NotTo(HaveOccurred())

			err = m.Wait()
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
