package logger

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Field", func() {
	Describe("MapToFields", func() {
		It("nil", func() {
			fields := MapToFields(nil)
			Expect(fields).To(BeNil())
		})

		It("one", func() {
			fields := MapToFields(map[string]any{
				"i": 1,
				"f": 1.1,
				"s": "str",
			})
			Expect(fields).To(ConsistOf(
				Field{Key: "i", Value: 1},
				Field{Key: "f", Value: 1.1},
				Field{Key: "s", Value: "str"},
			))
		})
	})
})
