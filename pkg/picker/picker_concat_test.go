package picker

import (
	stderrors "errors"

	"github.com/vesoft-inc/nebula-importer/v4/pkg/errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ConcatItems", func() {
	DescribeTable("types",
		func(item any, expectErr error) {
			ci := ConcatItems{}
			err := ci.Add(item)
			if expectErr == nil {
				Expect(err).NotTo(HaveOccurred())
			} else {
				Expect(err).To(HaveOccurred())
				Expect(stderrors.Is(err, expectErr)).To(BeTrue())
			}
		},
		EntryDescription("type %[1]T"),
		Entry(nil, uint8(1), nil),
		Entry(nil, int8(1), nil),
		Entry(nil, int8(-1), errors.ErrInvalidIndex),
		Entry(nil, uint16(1), nil),
		Entry(nil, int16(1), nil),
		Entry(nil, int16(-1), errors.ErrInvalidIndex),
		Entry(nil, uint32(1), nil),
		Entry(nil, int32(1), nil),
		Entry(nil, int32(-1), errors.ErrInvalidIndex),
		Entry(nil, uint64(1), nil),
		Entry(nil, int64(1), nil),
		Entry(nil, int64(-1), errors.ErrInvalidIndex),
		Entry(nil, uint(1), nil),
		Entry(nil, int(1), nil),
		Entry(nil, int(-1), errors.ErrInvalidIndex),
		Entry(nil, "str", nil),
		Entry(nil, []byte("str"), nil),
		Entry(nil, struct{}{}, errors.ErrUnsupportedConcatItemType),
	)

	It("nil", func() {
		ci := ConcatItems{}
		err := ci.Add()
		Expect(err).NotTo(HaveOccurred())
		Expect(ci.Len()).To(Equal(0))
	})

	It("many", func() {
		ci := ConcatItems{}
		err := ci.Add(1, "str1", 2, []byte("str2"), 3)
		Expect(err).NotTo(HaveOccurred())
		Expect(ci.Len()).To(Equal(5))
	})
})

var _ = Describe("ConcatPicker", func() {
	DescribeTable(".Pick",
		func(items []any, records []string, expectValue *Value, expectErr error) {
			ci := ConcatItems{}
			Expect(ci.Add(items...)).To(Not(HaveOccurred()))

			cp := ConcatPicker{
				items: ci,
			}
			value, err := cp.Pick(records)
			if expectErr != nil {
				Expect(err).To(HaveOccurred())
				Expect(stderrors.Is(err, expectErr)).To(BeTrue())
				Expect(value).To(BeNil())
			} else {
				Expect(err).NotTo(HaveOccurred())
				Expect(value).To(Equal(expectValue))
			}
		},
		Entry("int", []any{0}, []string{"10"}, &Value{Val: "10"}, nil),
		Entry("string", []any{"str"}, []string{"10"}, &Value{Val: "str"}, nil),
		Entry("mixed", []any{0, "str", 2}, []string{"10", "11", "12"}, &Value{Val: "10str12"}, nil),
		Entry("pick failed", []any{0, "str", 2}, []string{"10", "11"}, nil, errors.ErrNoRecord),
	)
})
