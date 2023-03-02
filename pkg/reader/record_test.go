package reader

import (
	"github.com/vesoft-inc/nebula-importer/v4/pkg/source"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("RecordReader", func() {
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
		Expect(s).NotTo(BeNil())
	})
	AfterEach(func() {
		err := s.Close()
		Expect(err).NotTo(HaveOccurred())
	})
	It("should success", func() {
		r := NewRecordReader(s)
		Expect(r).NotTo(BeNil())
	})
})
