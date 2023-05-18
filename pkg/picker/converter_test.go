package picker

import (
	stderrors "errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ConverterFunc", func() {
	It("normal Converters", func() {
		var converter Converter = ConverterFunc(func(v *Value) (*Value, error) {
			v.Val = "test " + v.Val
			return v, nil
		})
		v, err := converter.Convert(&Value{
			Val: "v",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(v).To(Equal(&Value{
			Val: "test v",
		}))
	})
})

var _ = Describe("Converters", func() {
	It("nil Converters", func() {
		var converter Converter = Converters(nil)
		v, err := converter.Convert(&Value{
			Val: "v",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(v).To(Equal(&Value{
			Val: "v",
		}))
	})

	It("one Converters", func() {
		converter := Converters{
			ConverterFunc(func(v *Value) (*Value, error) {
				v.Val = "test " + v.Val
				return v, nil
			}),
		}
		v, err := converter.Convert(&Value{
			Val: "v",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(v).To(Equal(&Value{
			Val: "test v",
		}))
	})

	It("many Converters", func() {
		converter := Converters{
			ConverterFunc(func(v *Value) (*Value, error) {
				v.Val = "test1 " + v.Val
				return v, nil
			}),
			ConverterFunc(func(v *Value) (*Value, error) {
				v.Val = "test2 " + v.Val
				return v, nil
			}),
			ConverterFunc(func(v *Value) (*Value, error) {
				v.Val = "test3 " + v.Val
				return v, nil
			}),
		}
		v, err := converter.Convert(&Value{
			Val: "v",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(v).To(Equal(&Value{
			Val: "test3 test2 test1 v",
		}))
	})

	It("many Converters failed", func() {
		converter := Converters{
			ConverterFunc(func(v *Value) (*Value, error) {
				v.Val = "test1 " + v.Val
				return v, nil
			}),
			ConverterFunc(func(v *Value) (*Value, error) {
				v.Val = "test2 " + v.Val
				return v, nil
			}),
			ConverterFunc(func(v *Value) (*Value, error) {
				return nil, stderrors.New("test failed")
			}),
		}
		v, err := converter.Convert(&Value{
			Val: "v",
		})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("test failed"))
		Expect(v).To(BeNil())
	})
})

var _ = Describe("NullableConverters", func() {
	It("nil NullableConverters", func() {
		converter := NullableConverters(nil)
		v, err := converter.Convert(&Value{
			Val: "v",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(v).To(Equal(&Value{
			Val: "v",
		}))

		v, err = converter.Convert(&Value{
			Val:       "v",
			IsNull:    true,
			isSetNull: true,
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(v).To(Equal(&Value{
			Val:       "v",
			IsNull:    true,
			isSetNull: true,
		}))
	})

	It("one NullableConverters", func() {
		converter := NullableConverters{
			ConverterFunc(func(v *Value) (*Value, error) {
				v.Val = "test " + v.Val
				return v, nil
			}),
		}
		v, err := converter.Convert(&Value{
			Val: "v",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(v).To(Equal(&Value{
			Val: "test v",
		}))
	})

	It("many NullableConverters", func() {
		converter := NullableConverters{
			ConverterFunc(func(v *Value) (*Value, error) {
				v.Val = "test1 " + v.Val
				return v, nil
			}),
			ConverterFunc(func(v *Value) (*Value, error) {
				v.Val = "test2 " + v.Val
				return v, nil
			}),
			ConverterFunc(func(v *Value) (*Value, error) {
				v.Val = "test3 " + v.Val
				return v, nil
			}),
		}
		v, err := converter.Convert(&Value{
			Val: "v",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(v).To(Equal(&Value{
			Val: "test3 test2 test1 v",
		}))
	})

	It("many NullableConverters failed", func() {
		converter := NullableConverters{
			ConverterFunc(func(v *Value) (*Value, error) {
				v.Val = "test1 " + v.Val
				return v, nil
			}),
			ConverterFunc(func(v *Value) (*Value, error) {
				v.Val = "test2 " + v.Val
				return v, nil
			}),
			ConverterFunc(func(v *Value) (*Value, error) {
				return nil, stderrors.New("test failed")
			}),
		}
		v, err := converter.Convert(&Value{
			Val: "v",
		})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("test failed"))
		Expect(v).To(BeNil())
	})

	It("many NullableConverters isSetNull", func() {
		converter := NullableConverters{
			ConverterFunc(func(v *Value) (*Value, error) {
				v.Val = "test1 " + v.Val
				return v, nil
			}),
			ConverterFunc(func(v *Value) (*Value, error) {
				v.Val = ""
				v.isSetNull = true
				return v, nil
			}),
			ConverterFunc(func(v *Value) (*Value, error) {
				return nil, stderrors.New("test failed")
			}),
		}
		v, err := converter.Convert(&Value{
			Val: "v",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(v).To(Equal(&Value{
			Val:       "",
			isSetNull: true,
		}))
	})
})
