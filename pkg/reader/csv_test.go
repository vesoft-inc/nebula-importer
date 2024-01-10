package reader

import (
	stderrors "errors"
	"io"

	"github.com/vesoft-inc/nebula-importer/v4/pkg/source"
	"github.com/vesoft-inc/nebula-importer/v4/pkg/spec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("csvReader", func() {
	Describe("default", func() {
		var s source.Source
		BeforeEach(func() {
			var err error
			s, err = source.New(&source.Config{
				Local: &source.LocalConfig{
					Path: "testdata/local.csv",
				},
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(s).NotTo(BeNil())
			err = s.Open()
			Expect(err).NotTo(HaveOccurred())
		})
		AfterEach(func() {
			err := s.Close()
			Expect(err).NotTo(HaveOccurred())
		})

		It("should success", func() {
			var (
				nBytes int64
				n      int
				record spec.Record
				err    error
			)
			r := NewCSVReader(s)
			nBytes, err = r.Size()
			Expect(err).NotTo(HaveOccurred())
			Expect(nBytes).To(Equal(int64(33)))

			n, record, err = r.Read()
			Expect(err).NotTo(HaveOccurred())
			Expect(n).To(Equal(6))
			Expect(record).To(Equal(spec.Record{"1", "2", "3"}))

			n, record, err = r.Read()
			Expect(err).NotTo(HaveOccurred())
			Expect(n).To(Equal(7))
			Expect(record).To(Equal(spec.Record{"4", " 5", "6"}))

			n, record, err = r.Read()
			Expect(err).NotTo(HaveOccurred())
			Expect(n).To(Equal(8))
			Expect(record).To(Equal(spec.Record{" 7", "8", " 9"}))

			n, record, err = r.Read()
			Expect(err).NotTo(HaveOccurred())
			Expect(n).To(Equal(12))
			Expect(record).To(Equal(spec.Record{"10", " 11 ", " 12"}))

			n, record, err = r.Read()
			Expect(err).To(HaveOccurred())
			Expect(stderrors.Is(err, io.EOF)).To(BeTrue())
			Expect(n).To(Equal(0))
			Expect(record).To(BeEmpty())
		})
	})

	Describe("with delimiter", func() {
		var s source.Source
		BeforeEach(func() {
			var err error
			s, err = source.New(&source.Config{
				Local: &source.LocalConfig{
					Path: "testdata/local_delimiter.csv",
				},
				CSV: &source.CSVConfig{
					Delimiter: "|",
					Comment:   "#",
				},
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(s).NotTo(BeNil())
			err = s.Open()
			Expect(err).NotTo(HaveOccurred())
		})
		AfterEach(func() {
			err := s.Close()
			Expect(err).NotTo(HaveOccurred())
		})

		It("should success", func() {
			var (
				nBytes int64
				n      int
				record spec.Record
				err    error
			)
			r := NewCSVReader(s)
			nBytes, err = r.Size()
			Expect(err).NotTo(HaveOccurred())
			Expect(nBytes).To(Equal(int64(117)))

			n, record, err = r.Read()
			Expect(err).NotTo(HaveOccurred())
			Expect(n).To(Equal(20))
			Expect(record).To(Equal(spec.Record{"1", "2", "3"}))

			n, record, err = r.Read()
			Expect(err).NotTo(HaveOccurred())
			Expect(n).To(Equal(35))
			Expect(record).To(Equal(spec.Record{"4", " 5", "6"}))

			n, record, err = r.Read()
			Expect(err).NotTo(HaveOccurred())
			Expect(n).To(Equal(22))
			Expect(record).To(Equal(spec.Record{" 7", "8", " 9"}))

			n, record, err = r.Read()
			Expect(err).NotTo(HaveOccurred())
			Expect(n).To(Equal(26))
			Expect(record).To(Equal(spec.Record{"10", " 11 ", " 12"}))

			n, record, err = r.Read()
			Expect(err).To(HaveOccurred())
			Expect(stderrors.Is(err, io.EOF)).To(BeTrue())
			Expect(n).To(Equal(14))
			Expect(record).To(BeEmpty())
		})
	})

	Describe("withHeader", func() {
		var s source.Source
		BeforeEach(func() {
			var err error
			s, err = source.New(&source.Config{
				Local: &source.LocalConfig{
					Path: "testdata/local_withHeader.csv",
				},
				CSV: &source.CSVConfig{
					WithHeader: true,
				},
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(s).NotTo(BeNil())
			err = s.Open()
			Expect(err).NotTo(HaveOccurred())
		})
		AfterEach(func() {
			err := s.Close()
			Expect(err).NotTo(HaveOccurred())
		})

		It("should success", func() {
			var (
				nBytes int64
				n      int
				record spec.Record
				err    error
			)
			r := NewCSVReader(s)
			nBytes, err = r.Size()
			Expect(err).NotTo(HaveOccurred())
			Expect(nBytes).To(Equal(int64(14)))

			n, record, err = r.Read()
			Expect(err).NotTo(HaveOccurred())
			Expect(n).To(Equal(14))
			Expect(record).To(Equal(spec.Record{"1", "2", "3"}))

			n, record, err = r.Read()
			Expect(err).To(HaveOccurred())
			Expect(stderrors.Is(err, io.EOF)).To(BeTrue())
			Expect(n).To(Equal(0))
			Expect(record).To(BeEmpty())
		})
	})

	Describe("withHeader read failed", func() {
		var s source.Source
		BeforeEach(func() {
			var err error
			s, err = source.New(&source.Config{
				Local: &source.LocalConfig{
					Path: "testdata/local_withHeader_failed.csv",
				},
				CSV: &source.CSVConfig{
					WithHeader: true,
				},
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(s).NotTo(BeNil())
			err = s.Open()
			Expect(err).NotTo(HaveOccurred())
		})
		AfterEach(func() {
			err := s.Close()
			Expect(err).NotTo(HaveOccurred())
		})

		It("should success", func() {
			var (
				nBytes int64
				err    error
			)
			r := NewCSVReader(s)
			nBytes, err = r.Size()
			Expect(err).NotTo(HaveOccurred())
			Expect(nBytes).To(Equal(int64(9)))

			_, _, err = r.Read()
			Expect(err).To(HaveOccurred())
		})
	})
})
