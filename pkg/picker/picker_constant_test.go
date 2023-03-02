package picker

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ConstantPicker", func() {
	It("normal ConstantPicker", func() {
		picker := ConstantPicker("test constant")

		v, err := picker.Pick([]string{})
		Expect(err).NotTo(HaveOccurred())
		Expect(v).To(Equal(&Value{
			Val: "test constant",
		}))
	})
})
