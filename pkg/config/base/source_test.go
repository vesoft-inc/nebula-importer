package configbase

import (
	stderrors "errors"

	"github.com/vesoft-inc/nebula-importer/v4/pkg/source"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Source", func() {
	Describe(".BuildSourceAndReader", func() {
		var (
			s          *Source
			ctrl       *gomock.Controller
			mockSource *source.MockSource
			patches    *gomonkey.Patches
		)
		BeforeEach(func() {
			ctrl = gomock.NewController(GinkgoT())
			mockSource = source.NewMockSource(ctrl)
			patches = gomonkey.NewPatches()
			s = &Source{
				SourceConfig: source.Config{
					Local: &source.LocalConfig{
						Path: "path",
					},
					CSV: &source.CSVConfig{
						Delimiter: ",",
					},
				},
				Batch: 7,
			}
		})
		AfterEach(func() {
			ctrl.Finish()
			patches.Reset()
		})
		It("successfully", func() {
			patches.ApplyGlobalVar(&sourceNew, func(_ *source.Config) (source.Source, error) {
				return mockSource, nil
			})

			mockSource.EXPECT().Name().AnyTimes().Return("source name")
			mockSource.EXPECT().Config().AnyTimes().Return(&s.SourceConfig)
			mockSource.EXPECT().Read(gomock.Any()).AnyTimes().DoAndReturn(func(p []byte) (int, error) {
				n := copy(p, "a,b,c\n")
				return n, nil
			})

			src, brr, err := s.BuildSourceAndReader()
			Expect(err).NotTo(HaveOccurred())
			Expect(src).NotTo(BeNil())
			Expect(brr).NotTo(BeNil())

			n, records, err := brr.ReadBatch()
			Expect(err).NotTo(HaveOccurred())
			Expect(records).To(HaveLen(7))
			Expect(n).To(Equal(6 * 7))
		})

		It("failed", func() {
			patches.ApplyGlobalVar(&sourceNew, func(_ *source.Config) (source.Source, error) {
				return nil, stderrors.New("test error")
			})
			src, brr, err := s.BuildSourceAndReader()
			Expect(err).To(HaveOccurred())
			Expect(src).To(BeNil())
			Expect(brr).To(BeNil())
		})
	})
})
