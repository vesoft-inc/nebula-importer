package reader

import (
	"github.com/vesoft-inc/nebula-importer/v4/pkg/logger"

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
			WithLogger(logger.NopLogger),
		)
		Expect(o).NotTo(BeNil())
		Expect(o.batch).To(Equal(100))
		Expect(o.logger).NotTo(BeNil())
	})
})
