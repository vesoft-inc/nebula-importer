package source

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("localSource", func() {
	It("exists", func() {
		s := newLocalSource(&Config{
			Local: &LocalConfig{
				Path: "testdata/local.txt",
			},
		})

		Expect(s.Name()).To(Equal("local testdata/local.txt"))

		err := s.Open()
		Expect(err).NotTo(HaveOccurred())
		Expect(s).NotTo(BeNil())

		Expect(s.Config()).NotTo(BeNil())

		nBytes, err := s.Size()
		Expect(err).NotTo(HaveOccurred())
		Expect(nBytes).To(Equal(int64(6)))

		var buf [1024]byte
		n, err := s.Read(buf[:])
		Expect(err).NotTo(HaveOccurred())
		Expect(n).To(Equal(6))

		err = s.Close()
		Expect(err).NotTo(HaveOccurred())
	})

	It("not exists", func() {
		s := newLocalSource(&Config{
			Local: &LocalConfig{
				Path: "testdata/not-exists.txt",
			},
		})
		err := s.Open()
		Expect(err).To(HaveOccurred())
	})

	It("get size failed", func() {
		s := newLocalSource(&Config{
			Local: &LocalConfig{
				Path: "testdata/local.txt",
			},
		})
		err := s.Open()
		Expect(err).NotTo(HaveOccurred())
		Expect(s).NotTo(BeNil())

		err = s.Close()
		Expect(err).NotTo(HaveOccurred())

		nBytes, err := s.Size()
		Expect(err).To(HaveOccurred())
		Expect(nBytes).To(Equal(int64(0)))
	})
})
