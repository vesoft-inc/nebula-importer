package specbase

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("StatementBuilderFunc", func() {
	It("", func() {
		var b StatementBuilder = StatementBuilderFunc(func(records ...Record) (string, error) {
			return "test statement", nil
		})
		statement, err := b.Build()
		Expect(err).NotTo(HaveOccurred())
		Expect(statement).To(Equal("test statement"))
	})
})
