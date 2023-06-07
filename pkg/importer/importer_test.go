package importer

import (
	stderrors "errors"
	"sync"
	"time"

	"github.com/vesoft-inc/nebula-importer/v4/pkg/client"
	"github.com/vesoft-inc/nebula-importer/v4/pkg/errors"
	"github.com/vesoft-inc/nebula-importer/v4/pkg/spec"
	specbase "github.com/vesoft-inc/nebula-importer/v4/pkg/spec/base"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Importer", func() {
	var (
		ctrl           *gomock.Controller
		mockClientPool *client.MockPool
		mockResponse   *client.MockResponse
		mockBuilder    *specbase.MockStatementBuilder
	)
	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockClientPool = client.NewMockPool(ctrl)
		mockResponse = client.NewMockResponse(ctrl)
		mockBuilder = specbase.NewMockStatementBuilder(ctrl)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("New", func() {
		It("build failed", func() {
			mockBuilder.EXPECT().Build(gomock.Any()).Return("", 0, errors.ErrNoRecord)

			i := New(mockBuilder, mockClientPool)
			resp, err := i.Import(spec.Record{})
			Expect(err).To(HaveOccurred())
			Expect(stderrors.Is(err, errors.ErrNoRecord)).To(BeTrue())
			Expect(resp).To(BeNil())
		})

		It("execute failed", func() {
			mockBuilder.EXPECT().Build(gomock.Any()).Return("statement", 1, nil)
			mockClientPool.EXPECT().Execute(gomock.Any()).Return(nil, stderrors.New("test error"))

			i := New(mockBuilder, mockClientPool)
			resp, err := i.Import(spec.Record{"id"})
			Expect(err).To(HaveOccurred())
			importError, ok := errors.AsImportError(err)
			Expect(ok).To(BeTrue())
			Expect(importError.Statement()).NotTo(BeEmpty())
			Expect(resp).To(BeNil())
		})

		It("execute IsSucceed false", func() {
			mockBuilder.EXPECT().Build(gomock.Any()).Return("statement", 1, nil)
			mockClientPool.EXPECT().Execute(gomock.Any()).Times(1).Return(mockResponse, nil)
			mockResponse.EXPECT().IsSucceed().Times(1).Return(false)
			mockResponse.EXPECT().GetError().Times(1).Return(stderrors.New("status failed"))

			i := New(mockBuilder, mockClientPool)
			resp, err := i.Import(spec.Record{"id"})
			Expect(err).To(HaveOccurred())
			importError, ok := errors.AsImportError(err)
			Expect(ok).To(BeTrue())
			Expect(importError.Messages).To(ContainElement(ContainSubstring("status failed")))
			Expect(importError.Statement()).NotTo(BeEmpty())
			Expect(resp).To(BeNil())
		})

		It("execute successfully", func() {
			mockBuilder.EXPECT().Build(gomock.Any()).Times(1).Return("statement", 1, nil)
			mockClientPool.EXPECT().Execute(gomock.Any()).Times(1).Return(mockResponse, nil)
			mockResponse.EXPECT().IsSucceed().Times(1).Return(true)
			mockResponse.EXPECT().GetLatency().Times(1).Return(time.Microsecond * 10)
			mockResponse.EXPECT().GetRespTime().AnyTimes().Return(time.Microsecond * 12)

			i := New(mockBuilder, mockClientPool)
			i.Wait()
			defer i.Done()
			resp, err := i.Import(spec.Record{"id"})
			Expect(err).NotTo(HaveOccurred())
			Expect(resp).NotTo(BeNil())
			Expect(resp.Latency).To(Equal(time.Microsecond * time.Duration(10)))
			Expect(resp.RespTime).To(Equal(time.Microsecond * time.Duration(12)))
		})

		It("execute successfully with Add, Wait and Done", func() {
			mockBuilder.EXPECT().Build(gomock.Any()).Times(2).Return("statement", 1, nil)
			mockClientPool.EXPECT().Execute(gomock.Any()).Times(2).Return(mockResponse, nil)
			mockResponse.EXPECT().IsSucceed().Times(2).Return(true)
			mockResponse.EXPECT().GetLatency().Times(2).Return(time.Microsecond * 10)
			mockResponse.EXPECT().GetRespTime().AnyTimes().Return(time.Microsecond * 12)

			var (
				wg              sync.WaitGroup
				isImporter1Done = false
			)
			// i2 need to wait i1
			i1 := New(mockBuilder, mockClientPool,
				WithAddFunc(func(delta int) {
					wg.Add(delta)
				}),
				WithDoneFunc(func() {
					time.Sleep(time.Millisecond)
					isImporter1Done = true
					wg.Done()
				}),
			)
			i2 := New(mockBuilder, mockClientPool,
				WithWaitFunc(func() {
					wg.Wait()
					Expect(isImporter1Done).To(BeTrue())
				}),
			)

			i1.Add(1)
			i2.Add(1)

			go func() {
				i1.Wait()
				defer i1.Done()
				resp, err := i1.Import(spec.Record{"id"})
				Expect(err).NotTo(HaveOccurred())
				Expect(resp).NotTo(BeNil())
				Expect(resp.Latency).To(Equal(time.Microsecond * time.Duration(10)))
			}()

			i2.Wait()
			defer i2.Done()
			resp, err := i2.Import(spec.Record{"id"})
			Expect(err).NotTo(HaveOccurred())
			Expect(resp).NotTo(BeNil())
			Expect(resp.Latency).To(Equal(time.Microsecond * time.Duration(10)))
			Expect(resp.RespTime).To(Equal(time.Microsecond * time.Duration(12)))
		})
	})
})
