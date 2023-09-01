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

	Describe(".Glob", func() {
		var (
			s           *Source
			ctrl        *gomock.Controller
			mockSource  *source.MockSource
			mockGlobber *source.MockGlobber
			patches     *gomonkey.Patches
		)
		BeforeEach(func() {
			ctrl = gomock.NewController(GinkgoT())
			mockSource = source.NewMockSource(ctrl)
			mockGlobber = source.NewMockGlobber(ctrl)
			patches = gomonkey.NewPatches()
			s = &Source{
				SourceConfig: source.Config{
					Local: &source.LocalConfig{
						Path: "path*",
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

		It("failed", func() {
			patches.ApplyGlobalVar(&sourceNew, func(_ *source.Config) (source.Source, error) {
				return nil, stderrors.New("test error")
			})
			ss, isSupportGlob, err := s.Glob()
			Expect(err).To(HaveOccurred())
			Expect(isSupportGlob).To(Equal(false))
			Expect(ss).To(BeNil())
		})

		It("unsupported", func() {
			patches.ApplyGlobalVar(&sourceNew, func(_ *source.Config) (source.Source, error) {
				return mockSource, nil
			})
			ss, isSupportGlob, err := s.Glob()
			Expect(err).NotTo(HaveOccurred())
			Expect(isSupportGlob).To(Equal(false))
			Expect(ss).To(BeNil())
		})

		It("failed at glob", func() {
			patches.ApplyGlobalVar(&sourceNew, func(_ *source.Config) (source.Source, error) {
				return struct {
					*source.MockSource
					*source.MockGlobber
				}{
					MockSource:  mockSource,
					MockGlobber: mockGlobber,
				}, nil
			})
			mockGlobber.EXPECT().Glob().Return(nil, stderrors.New("test error"))
			mockSource.EXPECT().Close().Return(nil)

			ss, isSupportGlob, err := s.Glob()
			Expect(err).To(HaveOccurred())
			Expect(isSupportGlob).To(Equal(true))
			Expect(ss).To(BeNil())
		})

		It("glob return empty", func() {
			patches.ApplyGlobalVar(&sourceNew, func(_ *source.Config) (source.Source, error) {
				return struct {
					*source.MockSource
					*source.MockGlobber
				}{
					MockSource:  mockSource,
					MockGlobber: mockGlobber,
				}, nil
			})
			mockGlobber.EXPECT().Glob().Return(nil, nil)
			mockSource.EXPECT().Name().AnyTimes().Return("source name")
			mockSource.EXPECT().Close().Return(nil)

			ss, isSupportGlob, err := s.Glob()
			Expect(err).To(HaveOccurred())
			Expect(isSupportGlob).To(Equal(true))
			Expect(ss).To(BeNil())
		})

		It("glob return many", func() {
			patches.ApplyGlobalVar(&sourceNew, func(_ *source.Config) (source.Source, error) {
				return struct {
					*source.MockSource
					*source.MockGlobber
				}{
					MockSource:  mockSource,
					MockGlobber: mockGlobber,
				}, nil
			})
			mockGlobber.EXPECT().Glob().Return([]*source.Config{
				{
					Local: &source.LocalConfig{
						Path: "path1",
					},
					CSV: &source.CSVConfig{
						Delimiter: ",",
					},
				},
				{
					Local: &source.LocalConfig{
						Path: "path2",
					},
					CSV: &source.CSVConfig{
						Delimiter: ",",
					},
				},
			}, nil)
			mockSource.EXPECT().Close().Return(nil)

			ss, isSupportGlob, err := s.Glob()
			Expect(err).NotTo(HaveOccurred())
			Expect(isSupportGlob).To(Equal(true))
			Expect(ss).To(Equal([]*Source{
				{
					SourceConfig: source.Config{
						Local: &source.LocalConfig{
							Path: "path1",
						},
						CSV: &source.CSVConfig{
							Delimiter: ",",
						},
					},
					Batch: 7,
				},
				{
					SourceConfig: source.Config{
						Local: &source.LocalConfig{
							Path: "path2",
						},
						CSV: &source.CSVConfig{
							Delimiter: ",",
						},
					},
					Batch: 7,
				},
			}))
		})
	})
})
