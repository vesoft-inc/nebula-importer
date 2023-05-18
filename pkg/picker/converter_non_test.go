package picker

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("NonConverter", func() {
	It("normal NonConverter", func() {
		var converter Converter = NonConverter{}

		v, err := converter.Convert(&Value{
			Val: "v",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(v).To(Equal(&Value{
			Val: "v",
		}))
	})
})
