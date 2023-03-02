package picker

import (
	stderrors "errors"

	"github.com/vesoft-inc/nebula-importer/v4/pkg/errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TypeConverter", func() {
	It("BOOL", func() {
		converter, _ := NewTypeConverter("BOOL")

		v, err := converter.Convert(&Value{
			Val: "true",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(v).To(Equal(&Value{
			Val: "true",
		}))

		v, err = converter.Convert(&Value{
			Val: "false",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(v).To(Equal(&Value{
			Val: "false",
		}))
	})

	It("INT", func() {
		converter, _ := NewTypeConverter("int")

		v, err := converter.Convert(&Value{
			Val: "0",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(v).To(Equal(&Value{
			Val: "0",
		}))
	})

	It("FLOAT", func() {
		converter, _ := NewTypeConverter("FLOAT")

		v, err := converter.Convert(&Value{
			Val: "1.2",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(v).To(Equal(&Value{
			Val: "1.2",
		}))
	})

	It("DOUBLE", func() {
		converter, _ := NewTypeConverter("DOUBLE")

		v, err := converter.Convert(&Value{
			Val: "1.2",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(v).To(Equal(&Value{
			Val: "1.2",
		}))
	})

	It("STRING", func() {
		converter, _ := NewTypeConverter("STRING")

		v, err := converter.Convert(&Value{
			Val: "str",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(v).To(Equal(&Value{
			Val: "\"str\"",
		}))
	})

	It("DATE", func() {
		converter, _ := NewTypeConverter("DATE")

		v, err := converter.Convert(&Value{
			Val: "2020-01-02",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(v).To(Equal(&Value{
			Val: "DATE(\"2020-01-02\")",
		}))
	})

	It("TIME", func() {
		converter, _ := NewTypeConverter("TIME")

		v, err := converter.Convert(&Value{
			Val: "18:38:23.284",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(v).To(Equal(&Value{
			Val: "TIME(\"18:38:23.284\")",
		}))
	})

	It("DATETIME", func() {
		converter, _ := NewTypeConverter("DATETIME")

		v, err := converter.Convert(&Value{
			Val: "2020-01-11T19:28:23.284",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(v).To(Equal(&Value{
			Val: "DATETIME(\"2020-01-11T19:28:23.284\")",
		}))
	})

	It("TIMESTAMP", func() {
		converter, _ := NewTypeConverter("TIMESTAMP")

		v, err := converter.Convert(&Value{
			Val: "2020-01-11T19:28:23",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(v).To(Equal(&Value{
			Val: "TIMESTAMP(\"2020-01-11T19:28:23\")",
		}))

		v, err = converter.Convert(&Value{
			Val: "1578770903",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(v).To(Equal(&Value{
			Val: "TIMESTAMP(1578770903)",
		}))
	})

	It("GEOGRAPHY", func() {
		converter, _ := NewTypeConverter("GEOGRAPHY")

		v, err := converter.Convert(&Value{
			Val: "Polygon((-85.1 34.8,-80.7 28.4,-76.9 34.9,-85.1 34.8))",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(v).To(Equal(&Value{
			Val: "ST_GeogFromText(\"Polygon((-85.1 34.8,-80.7 28.4,-76.9 34.9,-85.1 34.8))\")",
		}))
	})

	It("GEOGRAPHY(POINT)", func() {
		converter, _ := NewTypeConverter("GEOGRAPHY(POINT)")

		v, err := converter.Convert(&Value{
			Val: "Point(0.0 0.0)",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(v).To(Equal(&Value{
			Val: "ST_GeogFromText(\"Point(0.0 0.0)\")",
		}))
	})

	It("GEOGRAPHY(LINESTRING)", func() {
		converter, _ := NewTypeConverter("GEOGRAPHY(LINESTRING)")

		v, err := converter.Convert(&Value{
			Val: "linestring(0 1, 179.99 89.99)",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(v).To(Equal(&Value{
			Val: "ST_GeogFromText(\"linestring(0 1, 179.99 89.99)\")",
		}))
	})

	It("GEOGRAPHY(POLYGON)", func() {
		converter, _ := NewTypeConverter("GEOGRAPHY(POLYGON)")

		v, err := converter.Convert(&Value{
			Val: "polygon((0 1, 2 4, 3 5, 4 9, 0 1))",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(v).To(Equal(&Value{
			Val: "ST_GeogFromText(\"polygon((0 1, 2 4, 3 5, 4 9, 0 1))\")",
		}))
	})

	It("Unsupported", func() {
		converter, err := NewTypeConverter("Unsupported")

		Expect(err).To(HaveOccurred())
		Expect(stderrors.Is(err, errors.ErrUnsupportedValueType)).To(BeTrue())
		Expect(converter).To(BeNil())
	})
})
