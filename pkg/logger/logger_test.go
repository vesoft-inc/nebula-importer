package logger

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Level", func() {
	It("New", func() {
		l, err := New(WithConsole(false))
		Expect(err).NotTo(HaveOccurred())
		Expect(l).NotTo(BeNil())
		err = l.Sync()
		Expect(err).NotTo(HaveOccurred())
		err = l.Close()
		Expect(err).NotTo(HaveOccurred())
	})
	It("New failed", func() {
		l, err := New(WithConsole(false), WithFiles("not-exists/1.log"))
		Expect(err).To(HaveOccurred())
		Expect(l).To(BeNil())
	})
})
