package reader

import (
	"log/slog"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Option", func() {
	It("newOptions", func() {
		o := newOptions()
		Expect(o).NotTo(BeNil())

		Expect(o.logger).NotTo(BeNil())
	})

	It("withXXX", func() {
		o := newOptions(
			WithBatch(100),
			WithLogger(slog.Default()),
		)
		Expect(o).NotTo(BeNil())
		Expect(o.batch).To(Equal(100))
		Expect(o.logger).NotTo(BeNil())
	})
})
