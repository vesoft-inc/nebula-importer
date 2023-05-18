package client

import (
	stderrors "errors"
	"sync/atomic"
	"time"

	"github.com/vesoft-inc/nebula-importer/v4/pkg/errors"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Clientxxx", func() {
	It("NewClient", func() {
		c := NewClient(WithAddress("127.0.0.1:9669"))
		c1, ok := c.(*defaultClient)
		Expect(ok).To(BeTrue())
		Expect(c1).NotTo(BeNil())
		Expect(c1.addresses).To(Equal([]string{"127.0.0.1:9669"}))
	})

	Describe(".Open", func() {
		var (
			ctrl        *gomock.Controller
			mockSession *MockSession
		)

		BeforeEach(func() {
			ctrl = gomock.NewController(GinkgoT())
			mockSession = NewMockSession(ctrl)
		})

		AfterEach(func() {
			ctrl.Finish()
		})

		It("no addresses", func() {
			c := NewClient()
			err := c.Open()
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(errors.ErrNoAddresses))
		})

		It("empty address", func() {
			c := NewClient(WithAddress(""))
			err := c.Open()
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(errors.ErrInvalidAddress))
		})

		It("host empty", func() {
			c := NewClient(WithAddress(":9669"))
			err := c.Open()
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(errors.ErrInvalidAddress))
		})

		It("port is not a number", func() {
			c := NewClient(WithAddress("127.0.0.1:x"))
			err := c.Open()
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(errors.ErrInvalidAddress))
		})

		It("real nebula session", func() {
			c := NewClient(WithAddress("127.0.0.1:0"))
			err := c.Open()
			Expect(err).To(HaveOccurred())
		})

		It("open session failed", func() {
			c := NewClient(
				WithAddress("127.0.0.1:9669"),
				WithNewSessionFunc(func(_ HostAddress) Session {
					return mockSession
				}),
			)
			mockSession.EXPECT().Open().Return(stderrors.New("test open failed"))
			err := c.Open()
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(stderrors.New("test open failed")))
		})

		It("clientInitFunc failed", func() {
			c := NewClient(
				WithAddress("127.0.0.1:9669"),
				WithNewSessionFunc(func(_ HostAddress) Session {
					return mockSession
				}),
				WithClientInitFunc(func(client Client) error {
					return stderrors.New("test open failed")
				}),
			)
			mockSession.EXPECT().Open().Return(nil)
			mockSession.EXPECT().Close().Return(nil)
			err := c.Open()
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(stderrors.New("test open failed")))
		})

		It("successfully", func() {
			c := NewClient(
				WithAddress("127.0.0.1:9669"),
				WithNewSessionFunc(func(_ HostAddress) Session {
					return mockSession
				}),
				WithClientInitFunc(func(client Client) error {
					return nil
				}),
			)
			mockSession.EXPECT().Open().Return(nil)
			err := c.Open()
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe(".Execute", func() {
		var (
			c            Client
			ctrl         *gomock.Controller
			mockSession  *MockSession
			mockResponse *MockResponse
		)

		BeforeEach(func() {
			ctrl = gomock.NewController(GinkgoT())
			mockSession = NewMockSession(ctrl)
			mockResponse = NewMockResponse(ctrl)
			c = NewClient(
				WithAddress("127.0.0.1:9669"),
				WithRetryInitialInterval(time.Microsecond),
				WithNewSessionFunc(func(_ HostAddress) Session {
					return mockSession
				}),
			)

			mockSession.EXPECT().Open().Return(nil)
			err := c.Open()
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			ctrl.Finish()
		})

		It("retry case1", func() {
			// * Case 1: retry no more
			mockSession.EXPECT().Execute("test Execute statement").Times(1).Return(mockResponse, nil)
			mockResponse.EXPECT().IsSucceed().Times(1).Return(false)
			mockResponse.EXPECT().GetError().Times(1).Return(stderrors.New("test error"))
			mockResponse.EXPECT().IsPermanentError().Times(2).Return(true)

			resp, err := c.Execute("test Execute statement")
			Expect(err).NotTo(HaveOccurred())
			Expect(resp).NotTo(BeNil())
			Expect(resp.IsPermanentError()).To(BeTrue())
		})

		It("retry case2", func() {
			retryTimes := DefaultRetry + 10
			var currExecuteTimes int64
			fnIsSucceed := func() bool {
				curr := atomic.AddInt64(&currExecuteTimes, 1)
				return curr > int64(retryTimes)
			}

			// * Case 2. retry as much as possible
			mockSession.EXPECT().Execute("test Execute statement").Times(retryTimes+1).Return(mockResponse, nil)
			mockResponse.EXPECT().IsSucceed().Times(retryTimes + 2).DoAndReturn(fnIsSucceed)
			mockResponse.EXPECT().GetError().Times(retryTimes).Return(stderrors.New("test error"))
			mockResponse.EXPECT().IsPermanentError().Times(retryTimes).Return(false)
			mockResponse.EXPECT().IsRetryMoreError().Times(retryTimes).Return(true)

			resp, err := c.Execute("test Execute statement")
			Expect(err).NotTo(HaveOccurred())
			Expect(resp).NotTo(BeNil())
			Expect(resp.IsSucceed()).To(BeTrue())
		})

		It("retry case3", func() {
			// * Case 3: retry with limit times
			mockSession.EXPECT().Execute("test Execute statement").Times(DefaultRetry+1).Return(nil, stderrors.New("execute failed"))

			resp, err := c.Execute("test Execute statement")
			Expect(err).To(HaveOccurred())
			Expect(resp).To(BeNil())
		})

		It("successfully", func() {
			mockSession.EXPECT().Execute("test Execute statement").Times(1).Return(mockResponse, nil)
			mockResponse.EXPECT().IsSucceed().Times(1).Return(true)

			resp, err := c.Execute("test Execute statement")
			Expect(err).NotTo(HaveOccurred())
			Expect(resp).NotTo(BeNil())
		})
	})

	Describe(".Close", func() {
		var (
			c           Client
			ctrl        *gomock.Controller
			mockSession *MockSession
		)

		BeforeEach(func() {
			ctrl = gomock.NewController(GinkgoT())
			mockSession = NewMockSession(ctrl)
			c = NewClient(
				WithAddress("127.0.0.1:9669"),
				WithRetryInitialInterval(time.Microsecond),
				WithNewSessionFunc(func(_ HostAddress) Session {
					return mockSession
				}),
			)

			mockSession.EXPECT().Open().Return(nil)
			err := c.Open()
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			ctrl.Finish()
		})

		It("open session failed", func() {
			mockSession.EXPECT().Close().Return(stderrors.New("open session failed"))
			err := c.Close()
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(stderrors.New("open session failed")))
		})

		It("successfully", func() {
			mockSession.EXPECT().Close().Return(nil)
			err := c.Close()
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
