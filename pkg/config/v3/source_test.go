package configv3

import (
	"os"
	"path/filepath"

	"github.com/vesoft-inc/nebula-importer/v4/pkg/source"
	specv3 "github.com/vesoft-inc/nebula-importer/v4/pkg/spec/v3"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Source", func() {
	Describe(".BuildGraph", func() {
		It("Validate failed", func() {
			s := &Source{}
			graph, err := s.BuildGraph("")
			Expect(err).To(HaveOccurred())
			Expect(graph).To(BeNil())
		})

		It("successfully", func() {
			s := &Source{
				Nodes: specv3.Nodes{
					&specv3.Node{
						Name: "n1",
						ID: &specv3.NodeID{
							Name:  "id",
							Type:  specv3.ValueTypeString,
							Index: 0,
						},
					},
					&specv3.Node{
						Name: "n2",
						ID: &specv3.NodeID{
							Name:  "id",
							Type:  specv3.ValueTypeString,
							Index: 0,
						},
					},
				},
				Edges: specv3.Edges{
					&specv3.Edge{
						Name: "e1",
						Src: &specv3.EdgeNodeRef{
							Name: "n1",
							ID: &specv3.NodeID{
								Name:  "id",
								Type:  specv3.ValueTypeString,
								Index: 0,
							},
						},
						Dst: &specv3.EdgeNodeRef{
							Name: "n2",
							ID: &specv3.NodeID{
								Name:  "id",
								Type:  specv3.ValueTypeString,
								Index: 1,
							},
						},
					},
				},
			}
			graph, err := s.BuildGraph("graphName")
			Expect(err).NotTo(HaveOccurred())
			Expect(graph).NotTo(BeNil())
		})
	})

	Describe(".BuildImporters", func() {
		It("BuildGraph failed", func() {
			s := &Source{}
			importers, err := s.BuildImporters("", nil)
			Expect(err).To(HaveOccurred())
			Expect(importers).To(BeNil())
		})

		It("successfully", func() {
			s := &Source{
				Nodes: specv3.Nodes{
					&specv3.Node{
						Name: "n1",
						ID: &specv3.NodeID{
							Name:  "id",
							Type:  specv3.ValueTypeString,
							Index: 0,
						},
					},
					&specv3.Node{
						Name: "n2",
						ID: &specv3.NodeID{
							Name:  "id",
							Type:  specv3.ValueTypeString,
							Index: 1,
						},
					},
				},
				Edges: specv3.Edges{
					&specv3.Edge{
						Name: "e1",
						Src: &specv3.EdgeNodeRef{
							Name: "n1",
							ID: &specv3.NodeID{
								Name:  "id",
								Type:  specv3.ValueTypeString,
								Index: 0,
							},
						},
						Dst: &specv3.EdgeNodeRef{
							Name: "n2",
							ID: &specv3.NodeID{
								Name:  "id",
								Type:  specv3.ValueTypeString,
								Index: 1,
							},
						},
					},
				},
			}

			importers, err := s.BuildImporters("graphName", nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(importers).To(HaveLen(3))
		})
	})
})

var _ = Describe("Sources", func() {
	DescribeTable(".OptimizePath",
		func(configPath string, files, expectFiles []string) {
			var sources Sources
			if files != nil {
				sources = make(Sources, len(files))
			}
			for i, file := range files {
				sources[i].SourceConfig.Local = &source.LocalConfig{
					Path: file,
				}
			}
			Expect(sources.OptimizePath(configPath)).NotTo(HaveOccurred())
			var sourcePaths []string
			if sources != nil {
				sourcePaths = make([]string, len(sources))
				for i := range sources {
					sourcePaths[i] = sources[i].SourceConfig.Local.Path
				}
			}
			Expect(sourcePaths).To(Equal(expectFiles))
		},
		EntryDescription("%[1]s : %[2]v => %[3]v"),

		Entry(nil, "f.yaml", nil, nil),
		Entry(nil, "./f.yaml", []string{"1.csv"}, []string{"1.csv"}),
		Entry(nil, "f.yaml", []string{"1.csv", "2.csv"}, []string{"1.csv", "2.csv"}),
		Entry(nil, "./f.yaml", []string{"d10/1.csv", "./d20/2.csv"}, []string{"d10/1.csv", "d20/2.csv"}),

		Entry(nil, "./d1/f.yaml", nil, nil),
		Entry(nil, "d1/f.yaml", []string{"1.csv"}, []string{"d1/1.csv"}),
		Entry(nil, "./d1/f.yaml", []string{"1.csv", "2.csv"}, []string{"d1/1.csv", "d1/2.csv"}),
		Entry(nil, "d1/f.yaml", []string{"d10/1.csv", "./d20/2.csv"}, []string{"d1/d10/1.csv", "d1/d20/2.csv"}),

		Entry(nil, "./d1/f.yaml", nil, nil),
		Entry(nil, "d1/f.yaml", []string{"/1.csv"}, []string{"/1.csv"}),
		Entry(nil, "./d1/f.yaml", []string{"/1.csv", "/2.csv"}, []string{"/1.csv", "/2.csv"}),
		Entry(nil, "d1/f.yaml", []string{"/d10/1.csv", "/d20/2.csv"}, []string{"/d10/1.csv", "/d20/2.csv"}),
	)

	Describe(".OptimizePathWildCard", func() {
		var wd string
		BeforeEach(func() {
			var err error
			wd, err = os.Getwd()
			Expect(err).NotTo(HaveOccurred())
		})

		It("nil", func() {
			var sources Sources
			Expect(sources.OptimizePathWildCard()).NotTo(HaveOccurred())
		})

		It("rel:WildCard:yes", func() {
			sources := make(Sources, 1)
			sources[0].Source.SourceConfig.Local = &source.LocalConfig{
				Path: filepath.Join("testdata", "file*"),
			}
			Expect(sources.OptimizePathWildCard()).NotTo(HaveOccurred())
			if Expect(sources).To(HaveLen(3)) {
				Expect(sources[0].SourceConfig.Local.Path).To(Equal(filepath.Join("testdata", "file10")))
				Expect(sources[1].SourceConfig.Local.Path).To(Equal(filepath.Join("testdata", "file11")))
				Expect(sources[2].SourceConfig.Local.Path).To(Equal(filepath.Join("testdata", "file20")))
			}
		})

		It("rel:WildCard:no", func() {
			sources := make(Sources, 3)
			sources[0].Source.SourceConfig.Local = &source.LocalConfig{
				Path: filepath.Join("testdata", "file10"),
			}
			sources[1].Source.SourceConfig.Local = &source.LocalConfig{
				Path: filepath.Join("testdata", "file11"),
			}
			sources[2].Source.SourceConfig.Local = &source.LocalConfig{
				Path: filepath.Join("testdata", "file20"),
			}

			Expect(sources.OptimizePathWildCard()).NotTo(HaveOccurred())
			if Expect(sources).To(HaveLen(3)) {
				Expect(sources[0].SourceConfig.Local.Path).To(Equal(filepath.Join("testdata", "file10")))
				Expect(sources[1].SourceConfig.Local.Path).To(Equal(filepath.Join("testdata", "file11")))
				Expect(sources[2].SourceConfig.Local.Path).To(Equal(filepath.Join("testdata", "file20")))
			}
		})

		It("abs:WildCard:yes", func() {
			sources := make(Sources, 1)
			sources[0].SourceConfig.Local = &source.LocalConfig{
				Path: filepath.Join(wd, "testdata", "file*"),
			}
			Expect(sources.OptimizePathWildCard()).NotTo(HaveOccurred())
			if Expect(sources).To(HaveLen(3)) {
				Expect(sources[0].SourceConfig.Local.Path).To(Equal(filepath.Join(wd, "testdata", "file10")))
				Expect(sources[1].SourceConfig.Local.Path).To(Equal(filepath.Join(wd, "testdata", "file11")))
				Expect(sources[2].SourceConfig.Local.Path).To(Equal(filepath.Join(wd, "testdata", "file20")))
			}
		})

		It("abs:WildCard:no", func() {
			sources := make(Sources, 3)
			sources[0].Source.SourceConfig.Local = &source.LocalConfig{
				Path: filepath.Join(wd, "testdata", "file10"),
			}
			sources[1].Source.SourceConfig.Local = &source.LocalConfig{
				Path: filepath.Join(wd, "testdata", "file11"),
			}
			sources[2].Source.SourceConfig.Local = &source.LocalConfig{
				Path: filepath.Join(wd, "testdata", "file20"),
			}

			Expect(sources.OptimizePathWildCard()).NotTo(HaveOccurred())
			if Expect(sources).To(HaveLen(3)) {
				Expect(sources[0].SourceConfig.Local.Path).To(Equal(filepath.Join(wd, "testdata", "file10")))
				Expect(sources[1].SourceConfig.Local.Path).To(Equal(filepath.Join(wd, "testdata", "file11")))
				Expect(sources[2].SourceConfig.Local.Path).To(Equal(filepath.Join(wd, "testdata", "file20")))
			}
		})

		It("rel:WildCard:yes:s3", func() {
			sources := make(Sources, 2)
			sources[0].Source.SourceConfig.Local = &source.LocalConfig{
				Path: filepath.Join("testdata", "file*"),
			}
			sources[1].Source.SourceConfig.S3 = &source.S3Config{
				Bucket: "bucket",
			}
			Expect(sources.OptimizePathWildCard()).NotTo(HaveOccurred())
			if Expect(sources).To(HaveLen(4)) {
				Expect(sources[0].SourceConfig.Local.Path).To(Equal(filepath.Join("testdata", "file10")))
				Expect(sources[1].SourceConfig.Local.Path).To(Equal(filepath.Join("testdata", "file11")))
				Expect(sources[2].SourceConfig.Local.Path).To(Equal(filepath.Join("testdata", "file20")))
				Expect(sources[3].SourceConfig.S3.Bucket).To(Equal("bucket"))
			}
		})

		It("failed", func() {
			sources := make(Sources, 2)
			sources[0].Source.SourceConfig.Local = &source.LocalConfig{
				Path: filepath.Join("testdata", "file*"),
			}
			sources[1].SourceConfig.Local = &source.LocalConfig{
				Path: filepath.Join("testdata", "[a-b"),
			}
			Expect(sources.OptimizePathWildCard()).To(HaveOccurred())

			sources = make(Sources, 2)
			sources[0].Source.SourceConfig.Local = &source.LocalConfig{
				Path: filepath.Join("testdata", "file*"),
			}
			sources[1].SourceConfig.Local = &source.LocalConfig{
				Path: filepath.Join("testdata", "not-exists"),
			}
			Expect(sources.OptimizePathWildCard()).To(HaveOccurred())
		})
	})
})
