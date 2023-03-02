package util

import (
	stderrors "errors"
	"io"

	"github.com/agiledragon/gomonkey/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("CheckErr", func() {
	It("nil", func() {
		CheckErr(nil)
	})

	It("no import error", func() {
		patches := gomonkey.NewPatches()
		defer patches.Reset()

		var (
			isFprintCalled bool
			exitCode       int
		)
		patches.ApplyGlobalVar(&fnFprint, func(io.Writer, ...any) (int, error) {
			isFprintCalled = true
			return 0, nil
		})
		patches.ApplyGlobalVar(&fnExit, func(code int) {
			exitCode = code
		})

		CheckErr(stderrors.New("test error"))
		Expect(isFprintCalled).To(BeTrue())
		Expect(exitCode).To(Equal(1))
	})
})
