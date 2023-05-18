package picker

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("FunctionConverter", func() {
	It("normal FunctionConverter", func() {
		var converter Converter = FunctionConverter{
			Name: "testFunc",
		}

		v, err := converter.Convert(&Value{
			Val: "v",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(v).To(Equal(&Value{
			Val: "testFunc(v)",
		}))
	})
})

var _ = Describe("FunctionStringConverter", func() {
	It("normal FunctionStringConverter", func() {
		var converter Converter = FunctionStringConverter{
			Name: "testFunc",
		}

		v, err := converter.Convert(&Value{
			Val: "v",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(v).To(Equal(&Value{
			Val: "testFunc(\"v\")",
		}))
	})
})
