package specbase

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("StatementBuilderFunc", func() {
	It("", func() {
		var b StatementBuilder = StatementBuilderFunc(func(records ...Record) (string, int, error) {
			return "test statement", 1, nil
		})
		statement, nRecord, err := b.Build()
		Expect(err).NotTo(HaveOccurred())
		Expect(nRecord).To(Equal(1))
		Expect(statement).To(Equal("test statement"))
	})
})
