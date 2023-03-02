package client

import (
	stderrors "errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/vesoft-inc/nebula-importer/v4/pkg/errors"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Pool", func() {
	It("NewPool", func() {
		p := NewPool(WithAddress("127.0.0.1:9669"))
		p1, ok := p.(*defaultPool)
		Expect(ok).To(BeTrue())
		Expect(p1).NotTo(BeNil())
		Expect(p1.addresses).To(Equal([]string{"127.0.0.1:9669"}))
		Expect(p1.done).NotTo(BeNil())
		Expect(p1.chExecuteDataQueue).NotTo(BeNil())
	})

	Describe(".GetClient", func() {
		var (
			ctrl       *gomock.Controller
			mockClient *MockClient
		)

		BeforeEach(func() {
			ctrl = gomock.NewController(GinkgoT())
			mockClient = NewMockClient(ctrl)
		})

		AfterEach(func() {
			ctrl.Finish()
		})

		It("no addresses", func() {
			p := NewPool()
			c, err := p.GetClient()
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(errors.ErrNoAddresses))
			Expect(c).To(BeNil())
		})

		It("open client failed", func() {
			mockClient.EXPECT().Open().Return(stderrors.New("open client failed"))

			p := NewPool(
				WithAddress("127.0.0.1:9669", "127.0.0.2:9669"),
				func(o *options) {
					o.fnNewClientWithOptions = func(o *options) Client {
						Expect(o.addresses).To(Equal([]string{"127.0.0.1:9669"}))
						return mockClient
					}
				},
			)
			c, err := p.GetClient()
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(stderrors.New("open client failed")))
			Expect(c).To(BeNil())
		})

		It("successfully", func() {
			mockClient.EXPECT().Open().Return(nil)

			p := NewPool(
				WithAddress("127.0.0.1:9669", "127.0.0.2:9669"),
				func(o *options) {
					o.fnNewClientWithOptions = func(o *options) Client {
						Expect(o.addresses).To(Equal([]string{"127.0.0.1:9669"}))
						return mockClient
					}
				},
			)
			c, err := p.GetClient()
			Expect(err).NotTo(HaveOccurred())
			Expect(c).NotTo(BeNil())
		})
	})

	Describe(".Open", func() {
		var (
			ctrl       *gomock.Controller
			mockClient *MockClient
		)

		BeforeEach(func() {
			ctrl = gomock.NewController(GinkgoT())
			mockClient = NewMockClient(ctrl)
		})

		AfterEach(func() {
			ctrl.Finish()
		})

		It("no addresses", func() {
			pool := NewPool()
			err := pool.Open()
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(errors.ErrNoAddresses))
		})

		It("open Client failed", func() {
			pool := NewPool(
				WithAddress("127.0.0.1:9669"),
				func(o *options) {
					o.fnNewClientWithOptions = func(o *options) Client {
						return mockClient
					}
				},
			)

			mockClient.EXPECT().Open().Return(stderrors.New("open client failed"))

			err := pool.Open()
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(stderrors.New("open client failed")))
		})

		It("start workers successfully", func() {
			addresses := []string{"127.0.0.1:9669", "127.0.0.2:9669"}
			pool := NewPool(
				WithAddress(addresses...),
				func(o *options) {
					o.fnNewClientWithOptions = func(o *options) Client {
						return mockClient
					}
				},
			)

			var (
				// 1 for check and DefaultConcurrencyPerAddress for concurrency per address
				clientOpenTimes = (1 + DefaultConcurrencyPerAddress) * len(addresses)
				wg              sync.WaitGroup
			)

			wg.Add(clientOpenTimes)
			mockClient.EXPECT().Open().Times(clientOpenTimes).DoAndReturn(func() error {
				defer wg.Done()
				return nil
			})
			mockClient.EXPECT().Close().Times(clientOpenTimes).Return(nil)

			err := pool.Open()
			Expect(err).NotTo(HaveOccurred())

			wg.Wait()

			err = pool.Close()
			Expect(err).NotTo(HaveOccurred())

			resp, err := pool.Execute("test Execute statement")
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(ErrClosed))
			Expect(resp).To(BeNil())

			chExecuteResult, ok := pool.ExecuteChan("test ExecuteChan statement")
			Expect(ok).To(BeFalse())
			Expect(chExecuteResult).To(BeNil())
		})

		It("start workers temporary failure", func() {
			addresses := []string{"127.0.0.1:9669", "127.0.0.2:9669"}
			pool := NewPool(
				WithAddress(addresses...),
				WithReconnectInitialInterval(time.Nanosecond),
				func(o *options) {
					o.fnNewClientWithOptions = func(o *options) Client {
						return mockClient
					}
				},
			)

			var (
				// 1 for check and DefaultConcurrencyPerAddress for concurrency per address
				clientOpenTimes = (1 + DefaultConcurrencyPerAddress) * len(addresses)
				wg              sync.WaitGroup
				openTimes       int64
				failedOpenTimes = 10
			)

			wg.Add(clientOpenTimes + failedOpenTimes)
			fnOpen := func() error {
				defer wg.Done()
				curr := atomic.AddInt64(&openTimes, 1)
				if curr >= int64(1+DefaultConcurrencyPerAddress)+1 &&
					curr < int64(1+DefaultConcurrencyPerAddress)+1+int64(failedOpenTimes) {
					return stderrors.New("test start worker temporary failure")
				}
				return nil
			}

			mockClient.EXPECT().Open().Times(clientOpenTimes + failedOpenTimes).DoAndReturn(fnOpen)
			mockClient.EXPECT().Close().Times(clientOpenTimes).Return(nil)

			err := pool.Open()
			Expect(err).NotTo(HaveOccurred())

			wg.Wait()

			err = pool.Close()
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe(".Execute&.ExecuteChan", func() {
		var (
			ctrl         *gomock.Controller
			mockClient   *MockClient
			mockResponse *MockResponse
		)

		BeforeEach(func() {
			ctrl = gomock.NewController(GinkgoT())
			mockClient = NewMockClient(ctrl)
			mockResponse = NewMockResponse(ctrl)
		})

		AfterEach(func() {
			ctrl.Finish()
		})

		It("blocked at ExecuteChan", func() {
			addresses := []string{"127.0.0.1:9669"}
			pool := NewPool(
				WithAddress(addresses...),
				WithConcurrencyPerAddress(1),
				WithQueueSize(1),
				func(o *options) {
					o.fnNewClientWithOptions = func(o *options) Client {
						return mockClient
					}
				},
			)

			var (
				// 1 for check and DefaultConcurrencyPerAddress for concurrency per address
				clientOpenTimes = (1 + 1) * len(addresses)
				wg              sync.WaitGroup
				received        = make(chan struct{}, 10)
				filled          = make(chan struct{})
				waitToExec      = make(chan struct{})
			)

			wg.Add(clientOpenTimes)
			fnExecute := func(_ string) (Response, error) {
				received <- struct{}{}
				<-filled
				<-waitToExec
				return mockResponse, nil
			}

			mockClient.EXPECT().Open().Times(clientOpenTimes).DoAndReturn(func() error {
				defer wg.Done()
				return nil
			})
			mockClient.EXPECT().Execute("test ExecuteChan statement").MaxTimes(2).DoAndReturn(fnExecute)
			mockClient.EXPECT().Close().Times(clientOpenTimes).Return(nil)

			err := pool.Open()
			Expect(err).NotTo(HaveOccurred())

			// send request
			chExecuteResult, ok := pool.ExecuteChan("test ExecuteChan statement")
			Expect(ok).To(BeTrue())
			Expect(chExecuteResult).NotTo(BeNil())

			// already receive from chan
			<-received

			// fill up the chan
			chExecuteResult, ok = pool.ExecuteChan("test ExecuteChan statement")
			Expect(ok).To(BeTrue())
			Expect(chExecuteResult).NotTo(BeNil())
			close(filled)

			chExecuteResult, ok = pool.ExecuteChan("test ExecuteChan statement")
			Expect(ok).To(BeFalse())
			Expect(chExecuteResult).To(BeNil())

			// start to execute
			close(waitToExec)

			wg.Wait()

			err = pool.Close()
			Expect(err).NotTo(HaveOccurred())
		})

		It("chExecuteDataQueue closed", func() {
			addresses := []string{"127.0.0.1:9669", "127.0.0.2:9669"}
			pool := NewPool(
				WithAddress(addresses...),
				func(o *options) {
					o.fnNewClientWithOptions = func(o *options) Client {
						return mockClient
					}
				},
			)

			var (
				// 1 for check and DefaultConcurrencyPerAddress for concurrency per address
				clientOpenTimes = (1 + DefaultConcurrencyPerAddress) * len(addresses)
				wg              sync.WaitGroup
			)

			wg.Add(clientOpenTimes)

			mockClient.EXPECT().Open().Times(clientOpenTimes).DoAndReturn(func() error {
				defer wg.Done()
				return nil
			})
			mockClient.EXPECT().Close().Times(clientOpenTimes).Return(nil)

			err := pool.Open()
			Expect(err).NotTo(HaveOccurred())

			pool1 := pool.(*defaultPool)
			close(pool1.chExecuteDataQueue)

			wg.Wait()

			close(pool1.done)
			pool1.wgSession.Wait()
			Expect(err).NotTo(HaveOccurred())
		})

		It("concurrency", func() {
			var (
				addresses    = []string{"127.0.0.1:9669", "127.0.0.2:9669"}
				executeTimes = 1000
			)

			pool := NewPool(
				WithAddress(addresses...),
				WithQueueSize(executeTimes*2),
				func(o *options) {
					o.fnNewClientWithOptions = func(o *options) Client {
						return mockClient
					}
				},
			)

			var (
				// 1 for check and DefaultConcurrencyPerAddress for concurrency per address
				clientOpenTimes = (1 + DefaultConcurrencyPerAddress) * len(addresses)
				wg              sync.WaitGroup
			)
			wg.Add(clientOpenTimes)

			mockClient.EXPECT().Open().Times(clientOpenTimes).DoAndReturn(func() error {
				defer wg.Done()
				return nil
			})
			mockClient.EXPECT().Execute("test Execute statement").Times(executeTimes).Return(mockResponse, nil)
			mockClient.EXPECT().Execute("test ExecuteChan statement").Times(executeTimes).Return(mockResponse, nil)
			mockClient.EXPECT().Close().Times(clientOpenTimes).Return(nil)

			err := pool.Open()
			Expect(err).NotTo(HaveOccurred())

			var wgExecutes sync.WaitGroup
			for i := 0; i < executeTimes; i++ {
				wgExecutes.Add(1)
				go func() {
					defer GinkgoRecover()
					defer wgExecutes.Done()
					resp, err := pool.Execute("test Execute statement")
					Expect(err).NotTo(HaveOccurred())
					Expect(resp).NotTo(BeNil())
				}()

				wgExecutes.Add(1)
				go func() {
					defer GinkgoRecover()
					defer wgExecutes.Done()
					chExecuteResult, ok := pool.ExecuteChan("test ExecuteChan statement")
					Expect(ok).To(BeTrue())
					executeResult := <-chExecuteResult
					resp, err := executeResult.Response, executeResult.Err
					Expect(err).NotTo(HaveOccurred())
					Expect(resp).NotTo(BeNil())
				}()
			}
			wgExecutes.Wait()

			wg.Wait()

			err = pool.Close()
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
