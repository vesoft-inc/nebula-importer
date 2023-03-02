package util

import (
	stderrors "errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
)

var _ = Describe("Run", func() {
	It("success", func() {
		err := Run(&cobra.Command{})
		Expect(err).NotTo(HaveOccurred())
	})

	It("failed", func() {
		err := Run(&cobra.Command{
			RunE: func(_ *cobra.Command, _ []string) error {
				return stderrors.New("test error")
			},
		})
		Expect(err).To(HaveOccurred())
	})
})
