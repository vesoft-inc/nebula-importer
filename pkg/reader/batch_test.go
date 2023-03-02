package reader

import (
	stderrors "errors"
	"io"

	"github.com/vesoft-inc/nebula-importer/v4/pkg/source"
	"github.com/vesoft-inc/nebula-importer/v4/pkg/spec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	pkgerrors "github.com/pkg/errors"
)

var _ = Describe("BatchRecordReader", func() {
	When("successfully", func() {
		var (
			s  source.Source
			rr RecordReader
		)
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
			rr = NewRecordReader(s)
			Expect(rr).NotTo(BeNil())
		})
		AfterEach(func() {
			err := s.Close()
			Expect(err).NotTo(HaveOccurred())
		})
		It("default batch", func() {
			var (
				nBytes  int64
				n       int
				records []spec.Record
				err     error
			)
			brr := NewBatchRecordReader(rr, WithBatch(0))
			Expect(brr.Source()).NotTo(BeNil())
			nBytes, err = brr.Size()
			Expect(err).NotTo(HaveOccurred())
			Expect(nBytes).To(Equal(int64(33)))

			n, records, err = brr.ReadBatch()
			Expect(err).NotTo(HaveOccurred())
			Expect(n).To(Equal(33))
			Expect(records).To(Equal([]spec.Record{
				{"1", "2", "3"},
				{"4", " 5", "6"},
				{" 7", "8", " 9"},
				{"10", " 11 ", " 12"},
			}))

			n, records, err = brr.ReadBatch()
			Expect(err).To(HaveOccurred())
			Expect(stderrors.Is(err, io.EOF)).To(BeTrue())
			Expect(n).To(Equal(0))
			Expect(records).To(BeEmpty())
		})

		It("1 batch", func() {
			var (
				nBytes  int64
				n       int
				records []spec.Record
				err     error
			)
			brr := NewBatchRecordReader(rr, WithBatch(1))
			nBytes, err = brr.Size()
			Expect(err).NotTo(HaveOccurred())
			Expect(nBytes).To(Equal(int64(33)))

			n, records, err = brr.ReadBatch()
			Expect(err).NotTo(HaveOccurred())
			Expect(n).To(Equal(6))
			Expect(records).To(Equal([]spec.Record{
				{"1", "2", "3"},
			}))

			n, records, err = brr.ReadBatch()
			Expect(err).NotTo(HaveOccurred())
			Expect(n).To(Equal(7))
			Expect(records).To(Equal([]spec.Record{
				{"4", " 5", "6"},
			}))

			n, records, err = brr.ReadBatch()
			Expect(err).NotTo(HaveOccurred())
			Expect(n).To(Equal(8))
			Expect(records).To(Equal([]spec.Record{
				{" 7", "8", " 9"},
			}))

			n, records, err = brr.ReadBatch()
			Expect(err).NotTo(HaveOccurred())
			Expect(n).To(Equal(12))
			Expect(records).To(Equal([]spec.Record{
				{"10", " 11 ", " 12"},
			}))

			n, records, err = brr.ReadBatch()
			Expect(err).To(HaveOccurred())
			Expect(stderrors.Is(err, io.EOF)).To(BeTrue())
			Expect(n).To(Equal(0))
			Expect(records).To(BeEmpty())
		})

		It("2 batch", func() {
			var (
				nBytes  int64
				n       int
				records []spec.Record
				err     error
			)
			brr := NewBatchRecordReader(rr, WithBatch(1), WithBatch(2))
			nBytes, err = brr.Size()
			Expect(err).NotTo(HaveOccurred())
			Expect(nBytes).To(Equal(int64(33)))

			n, records, err = brr.ReadBatch()
			Expect(err).NotTo(HaveOccurred())
			Expect(n).To(Equal(13))
			Expect(records).To(Equal([]spec.Record{
				{"1", "2", "3"},
				{"4", " 5", "6"},
			}))
			n, records, err = brr.ReadBatch()
			Expect(err).NotTo(HaveOccurred())
			Expect(n).To(Equal(20))
			Expect(records).To(Equal([]spec.Record{
				{" 7", "8", " 9"},
				{"10", " 11 ", " 12"},
			}))

			n, records, err = brr.ReadBatch()
			Expect(err).To(HaveOccurred())
			Expect(stderrors.Is(err, io.EOF)).To(BeTrue())
			Expect(n).To(Equal(0))
			Expect(records).To(BeEmpty())
		})

		It("3 batch", func() {
			var (
				nBytes  int64
				n       int
				records []spec.Record
				err     error
			)
			brr := NewBatchRecordReader(rr, WithBatch(1), WithBatch(3))
			nBytes, err = brr.Size()
			Expect(err).NotTo(HaveOccurred())
			Expect(nBytes).To(Equal(int64(33)))

			n, records, err = brr.ReadBatch()
			Expect(err).NotTo(HaveOccurred())
			Expect(n).To(Equal(21))
			Expect(records).To(Equal([]spec.Record{
				{"1", "2", "3"},
				{"4", " 5", "6"},
				{" 7", "8", " 9"},
			}))
			n, records, err = brr.ReadBatch()
			Expect(err).NotTo(HaveOccurred())
			Expect(n).To(Equal(12))
			Expect(records).To(Equal([]spec.Record{
				{"10", " 11 ", " 12"},
			}))

			n, records, err = brr.ReadBatch()
			Expect(err).To(HaveOccurred())
			Expect(stderrors.Is(err, io.EOF)).To(BeTrue())
			Expect(n).To(Equal(0))
			Expect(records).To(BeEmpty())
		})

		It("4 batch", func() {
			var (
				nBytes  int64
				n       int
				records []spec.Record
				err     error
			)
			brr := NewBatchRecordReader(rr, WithBatch(1), WithBatch(4))
			nBytes, err = brr.Size()
			Expect(err).NotTo(HaveOccurred())
			Expect(nBytes).To(Equal(int64(33)))

			n, records, err = brr.ReadBatch()
			Expect(err).NotTo(HaveOccurred())
			Expect(n).To(Equal(33))
			Expect(records).To(Equal([]spec.Record{
				{"1", "2", "3"},
				{"4", " 5", "6"},
				{" 7", "8", " 9"},
				{"10", " 11 ", " 12"},
			}))

			n, records, err = brr.ReadBatch()
			Expect(err).To(HaveOccurred())
			Expect(stderrors.Is(err, io.EOF)).To(BeTrue())
			Expect(n).To(Equal(0))
			Expect(records).To(BeEmpty())
		})
	})

	When("failed", func() {
		var (
			s  source.Source
			rr RecordReader
		)
		BeforeEach(func() {
			var err error
			s, err = source.New(&source.Config{
				Local: &source.LocalConfig{
					Path: "testdata/local_failed.csv",
				},
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(s).NotTo(BeNil())
			err = s.Open()
			Expect(err).NotTo(HaveOccurred())
			rr = NewRecordReader(s)
			Expect(rr).NotTo(BeNil())
		})
		AfterEach(func() {
			err := s.Close()
			Expect(err).NotTo(HaveOccurred())
		})

		It("", func() {
			var (
				nBytes  int64
				n       int
				records []spec.Record
				err     error
			)
			brr := NewBatchRecordReader(rr, WithBatch(1), WithBatch(2))
			nBytes, err = brr.Size()
			Expect(err).NotTo(HaveOccurred())
			Expect(nBytes).To(Equal(int64(16)))

			n, records, err = brr.ReadBatch()
			Expect(err).NotTo(HaveOccurred())
			Expect(n).To(Equal(13))
			Expect(records).To(Equal([]spec.Record{
				{"id1"},
				{"id3"},
			}))

			n, records, err = brr.ReadBatch()
			Expect(err).NotTo(HaveOccurred())
			Expect(n).To(Equal(3))
			Expect(records).To(Equal([]spec.Record{
				{"id4"},
			}))

			n, records, err = brr.ReadBatch()
			Expect(err).To(HaveOccurred())
			Expect(stderrors.Is(err, io.EOF)).To(BeTrue())
			Expect(n).To(Equal(0))
			Expect(records).To(BeEmpty())
		})
	})
})

var _ = Describe("continueError", func() {
	It("", func() {
		var baseErr = stderrors.New("test error")
		err := NewContinueError(baseErr)
		Expect(err.Error()).To(Equal(baseErr.Error()))
		Expect(stderrors.Unwrap(err)).To(Equal(baseErr))
		Expect(pkgerrors.Cause(err)).To(Equal(baseErr))
	})
})
