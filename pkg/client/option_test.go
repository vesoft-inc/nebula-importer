package client

import (
	"github.com/vesoft-inc/nebula-importer/v4/pkg/logger"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Option", func() {
	It("newOptions", func() {
		o := newOptions()
		Expect(o).NotTo(BeNil())

		Expect(o.addresses).To(Equal([]string(nil)))
		Expect(o.user).To(Equal(DefaultUser))
		Expect(o.password).To(Equal(DefaultPassword))
		Expect(o.retry).To(Equal(DefaultRetry))
		Expect(o.retryInitialInterval).To(Equal(DefaultRetryInitialInterval))
		Expect(o.logger).NotTo(BeNil())
		Expect(o.fnNewSession).NotTo(BeNil())
		Expect(o.clientInitFunc).To(BeNil())
		Expect(o.reconnectInitialInterval).To(Equal(DefaultReconnectInitialInterval))
		Expect(o.concurrencyPerAddress).To(Equal(DefaultConcurrencyPerAddress))
		Expect(o.queueSize).To(Equal(DefaultQueueSize))
		Expect(o.fnNewClientWithOptions).NotTo(BeNil())

		o1 := o.clone()
		Expect(o1.addresses).To(Equal([]string(nil)))
		Expect(o1.user).To(Equal(DefaultUser))
		Expect(o1.password).To(Equal(DefaultPassword))
		Expect(o1.retry).To(Equal(DefaultRetry))
		Expect(o1.retryInitialInterval).To(Equal(DefaultRetryInitialInterval))
		Expect(o1.logger).NotTo(BeNil())
		Expect(o1.fnNewSession).NotTo(BeNil())
		Expect(o1.clientInitFunc).To(BeNil())
		Expect(o1.reconnectInitialInterval).To(Equal(DefaultReconnectInitialInterval))
		Expect(o1.concurrencyPerAddress).To(Equal(DefaultConcurrencyPerAddress))
		Expect(o1.queueSize).To(Equal(DefaultQueueSize))
		Expect(o.fnNewClientWithOptions).NotTo(BeNil())

		o1.addresses = []string{"127.0.0.1:9669"}
		Expect(o.addresses).To(Equal([]string(nil)))
		Expect(o1.addresses).To(Equal([]string{"127.0.0.1:9669"}))
	})

	It("withXXX", func() {
		o := newOptions(
			WithV3(),
			WithAddress("127.0.0.1:9669"),
			WithAddress("127.0.0.2:9669,127.0.0.3:9669"),
			WithAddress("127.0.0.4:9669,127.0.0.5:9669", "127.0.0.6:9669"),
			WithUser("u0"),
			WithPassword("p0"),
			WithUserPassword("newUser", "newPassword"),
			WithRetry(DefaultRetry-1),
			WithRetry(DefaultRetry+1),
			WithRetryInitialInterval(DefaultRetryInitialInterval-1),
			WithRetryInitialInterval(DefaultRetryInitialInterval+1),
			WithLogger(logger.NopLogger),
			WithNewSessionFunc(func(HostAddress) Session { return nil }),
			WithClientInitFunc(func(Client) error { return nil }),
			WithReconnectInitialInterval(DefaultReconnectInitialInterval-1),
			WithReconnectInitialInterval(DefaultReconnectInitialInterval+1),
			WithConcurrencyPerAddress(DefaultConcurrencyPerAddress-1),
			WithConcurrencyPerAddress(DefaultConcurrencyPerAddress+1),
			WithQueueSize(DefaultQueueSize-1),
			WithQueueSize(DefaultQueueSize+1),
		)
		Expect(o).NotTo(BeNil())
		Expect(o.addresses).To(Equal([]string{
			"127.0.0.1:9669",
			"127.0.0.2:9669",
			"127.0.0.3:9669",
			"127.0.0.4:9669",
			"127.0.0.5:9669",
			"127.0.0.6:9669",
		}))
		Expect(o.user).To(Equal("newUser"))
		Expect(o.password).To(Equal("newPassword"))
		Expect(o.retry).To(Equal(DefaultRetry + 1))
		Expect(o.retryInitialInterval).To(Equal(DefaultRetryInitialInterval + 1))
		Expect(o.logger).NotTo(BeNil())
		Expect(o.fnNewSession).NotTo(BeNil())
		Expect(o.clientInitFunc).NotTo(BeNil())
		Expect(o.reconnectInitialInterval).To(Equal(DefaultReconnectInitialInterval + 1))
		Expect(o.concurrencyPerAddress).To(Equal(DefaultConcurrencyPerAddress + 1))
		Expect(o.queueSize).To(Equal(DefaultQueueSize + 1))
		Expect(o.fnNewClientWithOptions).NotTo(BeNil())

		o1 := o.clone()
		Expect(o1).NotTo(BeNil())
		Expect(o1.addresses).To(Equal([]string{
			"127.0.0.1:9669",
			"127.0.0.2:9669",
			"127.0.0.3:9669",
			"127.0.0.4:9669",
			"127.0.0.5:9669",
			"127.0.0.6:9669",
		}))
		Expect(o1.user).To(Equal("newUser"))
		Expect(o1.password).To(Equal("newPassword"))
		Expect(o1.retry).To(Equal(DefaultRetry + 1))
		Expect(o1.retryInitialInterval).To(Equal(DefaultRetryInitialInterval + 1))
		Expect(o1.logger).NotTo(BeNil())
		Expect(o1.fnNewSession).NotTo(BeNil())
		Expect(o1.clientInitFunc).NotTo(BeNil())
		Expect(o1.reconnectInitialInterval).To(Equal(DefaultReconnectInitialInterval + 1))
		Expect(o1.concurrencyPerAddress).To(Equal(DefaultConcurrencyPerAddress + 1))
		Expect(o1.queueSize).To(Equal(DefaultQueueSize + 1))
		Expect(o.fnNewClientWithOptions).NotTo(BeNil())

		o1.addresses = []string{"127.0.0.1:9669"}
		Expect(o.addresses).To(Equal([]string{
			"127.0.0.1:9669",
			"127.0.0.2:9669",
			"127.0.0.3:9669",
			"127.0.0.4:9669",
			"127.0.0.5:9669",
			"127.0.0.6:9669",
		}))
		Expect(o1.addresses).To(Equal([]string{"127.0.0.1:9669"}))
	})

	It("fnNewSession v3", func() {
		o := newOptions(WithV3())
		s := o.fnNewSession(HostAddress{})
		_, ok := s.(*defaultSessionV3)
		Expect(ok).To(BeTrue())
	})
})
