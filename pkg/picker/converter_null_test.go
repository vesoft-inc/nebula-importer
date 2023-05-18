package picker

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("NullConverter", func() {
	It("normal NullConverter", func() {
		var converter Converter = NullConverter{
			Value: "NULL",
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
			Val:       "NULL",
			IsNull:    true,
			isSetNull: true,
		}))
	})
})

var _ = Describe("NullableConverter", func() {
	It("normal NullableConverter", func() {
		var converter Converter = NullableConverter{
			Nullable: func(s string) bool {
				return s == "NULL"
			},
		}

		v, err := converter.Convert(&Value{
			Val: "v",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(v).To(Equal(&Value{
			Val: "v",
		}))

		v, err = converter.Convert(&Value{
			Val: "NULL",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(v).To(Equal(&Value{
			Val:    "NULL",
			IsNull: true,
		}))
	})
})
