package specbase

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Filter", func() {
	It("build failed", func() {
		f := Filter{
			Expr: "",
		}
		err := f.Build()
		Expect(err).To(HaveOccurred())
		Expect(f.program).To(BeNil())
	})

	It("successfully", func() {
		f := Filter{
			Expr: `(Record[0] == "A" or Record[0] == "B") and Record[1] != "C"`,
		}
		err := f.Build()
		Expect(err).NotTo(HaveOccurred())
		Expect(f.program).NotTo(BeNil())

		ok, err := f.Filter(Record{})
		Expect(err).To(HaveOccurred())
		Expect(ok).To(BeFalse())

		ok, err = f.Filter(Record{"A", "C"})
		Expect(err).NotTo(HaveOccurred())
		Expect(ok).To(BeFalse())

		ok, err = f.Filter(Record{"B", "D"})
		Expect(err).NotTo(HaveOccurred())
		Expect(ok).To(BeTrue())
	})
})
