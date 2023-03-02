package specv3

import (
	stderrors "errors"

	"github.com/vesoft-inc/nebula-importer/v4/pkg/errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Rank", func() {
	It(".Complete", func() {
		prop := &Rank{}
		prop.Complete()
	})

	DescribeTable(".Validate",
		func(rank *Rank, expectErr error) {
			err := rank.Validate()
			if expectErr != nil {
				if Expect(err).To(HaveOccurred()) {
					Expect(stderrors.Is(err, expectErr)).To(BeTrue())
					e, ok := errors.AsImportError(err)
					Expect(ok).To(BeTrue())
					Expect(e.Cause()).To(Equal(expectErr))
				}
			} else {
				Expect(err).NotTo(HaveOccurred())
			}
		},
		Entry("", &Rank{Index: -1}, errors.ErrInvalidIndex),
		Entry("normal", &Rank{Index: 0}, nil),
	)

	It(".Validate init picker failed", func() {
		rank := Rank{Index: -1}

		err := rank.Validate()
		Expect(err).To(HaveOccurred())
		Expect(stderrors.Is(err, errors.ErrInvalidIndex)).To(BeTrue())
	})

	DescribeTable(".Value",
		func(rank *Rank, record Record, expectValue string, expectErr error) {
			val, err := func() (string, error) {
				rank.Complete()
				err := rank.Validate()
				if err != nil {
					return "", err
				}
				return rank.Value(record)
			}()
			if expectErr != nil {
				if Expect(err).To(HaveOccurred()) {
					Expect(stderrors.Is(err, expectErr)).To(BeTrue())
					e, ok := errors.AsImportError(err)
					Expect(ok).To(BeTrue())
					Expect(e.Cause()).To(Equal(expectErr))
				}
				Expect(val).To(Equal(expectValue))
			} else {
				Expect(err).NotTo(HaveOccurred())
				Expect(val).To(Equal(expectValue))
			}
		},
		Entry("no record empty",
			&Rank{Index: 0},
			Record([]string{}),
			"",
			errors.ErrNoRecord,
		),
		Entry("no record",
			&Rank{Index: 1},
			Record([]string{"0"}),
			"",
			errors.ErrNoRecord,
		),
		Entry("successfully",
			&Rank{Index: 1},
			Record([]string{"1", "11"}),
			"11",
			nil,
		),
	)
})
