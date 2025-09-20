package configbase

import (
	"os"
	"path/filepath"

	"github.com/vesoft-inc/nebula-importer/v4/pkg/logger"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Log", func() {
	Describe(".BuildLogger", func() {
		var tmpdir string

		BeforeEach(func() {
			var err error
			tmpdir, err = os.MkdirTemp("", "test")
			Expect(err).NotTo(HaveOccurred())
		})
		AfterEach(func() {
			err := os.RemoveAll(tmpdir)
			Expect(err).NotTo(HaveOccurred())
		})
		It("failed", func() {
			var (
				level   = "INFO"
				console = true
			)
			configLog := Log{
				Level:   &level,
				Console: &console,
				Files:   []string{filepath.Join(tmpdir, "not-exists", "1.log")},
				Fields:  nil,
			}
			l, err := configLog.BuildLogger()
			Expect(err).To(HaveOccurred())
			Expect(l).To(BeNil())
		})

		It("success", func() {
			var (
				level   = "INFO"
				console = true
			)
			configLog := Log{
				Level:   &level,
				Console: &console,
				Files:   []string{filepath.Join(tmpdir, "1.log")},
				Fields:  logger.Fields{{Key: "k1", Value: "v1"}},
			}

			l, err := configLog.BuildLogger()
			Expect(err).NotTo(HaveOccurred())
			defer l.Close()
			Expect(l).NotTo(BeNil())
		})
	})

	It(".OptimizePath nil", func() {
		var configLog *Log
		Expect(configLog.OptimizePath("")).NotTo(HaveOccurred())
	})

	DescribeTable(".OptimizePath",
		func(configPath string, files, expectFiles []string) {
			l := &Log{
				Files: files,
			}
			Expect(l.OptimizePath(configPath)).NotTo(HaveOccurred())
			Expect(l.Files).To(Equal(expectFiles))
		},
		EntryDescription("%[1]s : %[2]v => %[3]v"),

		Entry(nil, "f.yaml", nil, nil),
		Entry(nil, "./f.yaml", []string{"1.log"}, []string{"1.log"}),
		Entry(nil, "f.yaml", []string{"1.log", "2.log"}, []string{"1.log", "2.log"}),
		Entry(nil, "./f.yaml", []string{"d10/1.log", "./d20/2.log"}, []string{"d10/1.log", "d20/2.log"}),

		Entry(nil, "./d1/f.yaml", nil, nil),
		Entry(nil, "d1/f.yaml", []string{"1.log"}, []string{"d1/1.log"}),
		Entry(nil, "./d1/f.yaml", []string{"1.log", "2.log"}, []string{"d1/1.log", "d1/2.log"}),
		Entry(nil, "d1/f.yaml", []string{"d10/1.log", "./d20/2.log"}, []string{"d1/d10/1.log", "d1/d20/2.log"}),

		Entry(nil, "./d1/f.yaml", nil, nil),
		Entry(nil, "d1/f.yaml", []string{"/1.log"}, []string{"/1.log"}),
		Entry(nil, "./d1/f.yaml", []string{"/1.log", "/2.log"}, []string{"/1.log", "/2.log"}),
		Entry(nil, "d1/f.yaml", []string{"/d10/1.log", "/d20/2.log"}, []string{"/d10/1.log", "/d20/2.log"}),
	)
})
