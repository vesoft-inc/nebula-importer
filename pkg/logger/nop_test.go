package logger

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("nopLogger", func() {
	It("nopLogger", func() {
		var (
			l   Logger
			err error
		)
		l = nopLogger{}
		l = l.SkipCaller(1).With().WithError(nil)
		l.Debug("")
		l.Info("")
		l.Warn("")
		l.Error("")
		l.Panic("")
		l.Fatal("")
		Expect(l).NotTo(BeNil())
		err = l.Sync()
		Expect(err).NotTo(HaveOccurred())
		err = l.Close()
		Expect(err).NotTo(HaveOccurred())
	})
})
