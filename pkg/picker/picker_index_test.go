package picker

import (
	stderrors "errors"

	"github.com/vesoft-inc/nebula-importer/v4/pkg/errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("IndexPicker", func() {
	It("normal IndexPicker", func() {
		picker := IndexPicker(1)

		v, err := picker.Pick(nil)
		Expect(err).To(HaveOccurred())
		Expect(stderrors.Is(err, errors.ErrNoRecord)).To(BeTrue())
		Expect(v).To(BeNil())

		v, err = picker.Pick([]string{})
		Expect(err).To(HaveOccurred())
		Expect(stderrors.Is(err, errors.ErrNoRecord)).To(BeTrue())
		Expect(v).To(BeNil())

		v, err = picker.Pick([]string{"v0"})
		Expect(err).To(HaveOccurred())
		Expect(stderrors.Is(err, errors.ErrNoRecord)).To(BeTrue())
		Expect(v).To(BeNil())

		v, err = picker.Pick([]string{"v0", "v1"})
		Expect(err).NotTo(HaveOccurred())
		Expect(v).To(Equal(&Value{
			Val: "v1",
		}))

		v, err = picker.Pick([]string{"v0", "v1", "v2"})
		Expect(err).NotTo(HaveOccurred())
		Expect(v).To(Equal(&Value{
			Val: "v1",
		}))
	})
})
