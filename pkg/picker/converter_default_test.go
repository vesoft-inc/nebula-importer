package picker

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("DefaultConverter", func() {
	It("normal DefaultConverter", func() {
		var converter Converter = DefaultConverter{
			Value: "default",
		}

		v, err := converter.Convert(&Value{
			Val: "v",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(v).To(Equal(&Value{
			Val: "v",
		}))

		v, err = converter.Convert(&Value{
			Val:    "v",
			IsNull: true,
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(v).To(Equal(&Value{
			Val:    "default",
			IsNull: false,
		}))
	})
})
