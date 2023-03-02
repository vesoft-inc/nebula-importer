package picker

import (
	stderrors "errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ConverterPicker", func() {
	It("normal ConverterPicker", func() {
		picker := ConverterPicker{
			picker: PickerFunc(func(strings []string) (*Value, error) {
				return &Value{Val: "v"}, nil
			}),
			converter: ConverterFunc(func(v *Value) (*Value, error) {
				v.Val = "test " + v.Val
				return v, nil
			}),
		}

		v, err := picker.Pick([]string{})
		Expect(err).NotTo(HaveOccurred())
		Expect(v).To(Equal(&Value{
			Val: "test v",
		}))
	})

	It("pick failed ConverterPicker", func() {
		picker := ConverterPicker{
			picker: PickerFunc(func(strings []string) (*Value, error) {
				return nil, stderrors.New("test error")
			}),
			converter: ConverterFunc(func(v *Value) (*Value, error) {
				v.Val = "test " + v.Val
				return v, nil
			}),
		}

		v, err := picker.Pick([]string{})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("test error"))
		Expect(v).To(BeNil())
	})

	It("converter failed ConverterPicker", func() {
		picker := ConverterPicker{
			picker: PickerFunc(func(strings []string) (*Value, error) {
				return &Value{Val: "v"}, nil
			}),
			converter: ConverterFunc(func(v *Value) (*Value, error) {
				return nil, stderrors.New("test error")
			}),
		}

		v, err := picker.Pick([]string{})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("test error"))
		Expect(v).To(BeNil())
	})

	It("converter nil ConverterPicker", func() {
		picker := ConverterPicker{
			picker: PickerFunc(func(strings []string) (*Value, error) {
				return &Value{Val: "v"}, nil
			}),
		}

		v, err := picker.Pick([]string{})
		Expect(err).NotTo(HaveOccurred())
		Expect(v).To(Equal(&Value{
			Val: "v",
		}))
	})
})

var _ = Describe("NullablePickers", func() {
	It("one NullablePickers", func() {
		picker := NullablePickers{
			PickerFunc(func([]string) (*Value, error) {
				return &Value{Val: "v"}, nil
			}),
		}
		v, err := picker.Pick([]string{})
		Expect(err).NotTo(HaveOccurred())
		Expect(v).To(Equal(&Value{
			Val: "v",
		}))
	})

	It("one NullablePickers failed", func() {
		picker := NullablePickers{
			PickerFunc(func([]string) (*Value, error) {
				return nil, stderrors.New("test failed")
			}),
		}
		v, err := picker.Pick([]string{})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("test failed"))
		Expect(v).To(BeNil())
	})

	It("many NullablePickers first", func() {
		picker := NullablePickers{
			PickerFunc(func([]string) (*Value, error) {
				return &Value{Val: "v1"}, nil
			}),
			PickerFunc(func([]string) (*Value, error) {
				return &Value{Val: "v2"}, nil
			}),
			PickerFunc(func([]string) (*Value, error) {
				return &Value{Val: "v3"}, nil
			}),
		}
		v, err := picker.Pick([]string{})
		Expect(err).NotTo(HaveOccurred())
		Expect(v).To(Equal(&Value{
			Val: "v1",
		}))
	})

	It("many NullablePickers middle", func() {
		picker := NullablePickers{
			PickerFunc(func([]string) (*Value, error) {
				return &Value{IsNull: true}, nil
			}),
			PickerFunc(func([]string) (*Value, error) {
				return &Value{Val: "v2"}, nil
			}),
			PickerFunc(func([]string) (*Value, error) {
				return &Value{Val: "v3"}, nil
			}),
		}
		v, err := picker.Pick([]string{})
		Expect(err).NotTo(HaveOccurred())
		Expect(v).To(Equal(&Value{
			Val: "v2",
		}))
	})

	It("many NullablePickers last", func() {
		picker := NullablePickers{
			PickerFunc(func([]string) (*Value, error) {
				return &Value{IsNull: true}, nil
			}),
			PickerFunc(func([]string) (*Value, error) {
				return &Value{IsNull: true}, nil
			}),
			PickerFunc(func([]string) (*Value, error) {
				return &Value{Val: "v3"}, nil
			}),
		}
		v, err := picker.Pick([]string{})
		Expect(err).NotTo(HaveOccurred())
		Expect(v).To(Equal(&Value{
			Val: "v3",
		}))
	})

	It("many NullablePickers no", func() {
		picker := NullablePickers{
			PickerFunc(func([]string) (*Value, error) {
				return &Value{IsNull: true}, nil
			}),
			PickerFunc(func([]string) (*Value, error) {
				return &Value{IsNull: true}, nil
			}),
			PickerFunc(func([]string) (*Value, error) {
				return &Value{IsNull: true}, nil
			}),
		}
		v, err := picker.Pick([]string{})
		Expect(err).NotTo(HaveOccurred())
		Expect(v).To(Equal(&Value{
			Val:    "",
			IsNull: true,
		}))
	})

	It("many NullablePickers failed", func() {
		picker := NullablePickers{
			PickerFunc(func([]string) (*Value, error) {
				return &Value{IsNull: true}, nil
			}),
			PickerFunc(func([]string) (*Value, error) {
				return &Value{IsNull: true}, nil
			}),
			PickerFunc(func([]string) (*Value, error) {
				return nil, stderrors.New("test error")
			}),
		}
		v, err := picker.Pick([]string{})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("test error"))
		Expect(v).To(BeNil())
	})
})
