package configbase

import (
	stderrors "errors"
	"os"
	"time"

	"github.com/vesoft-inc/nebula-importer/v4/pkg/errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Client", func() {
	Describe(".BuildClientPool", func() {
		var (
			tmpdir string
		)
		BeforeEach(func() {
			var err error
			tmpdir, err = os.MkdirTemp("", "test")
			Expect(err).NotTo(HaveOccurred())
		})
		AfterEach(func() {
			err := os.RemoveAll(tmpdir)
			Expect(err).NotTo(HaveOccurred())
		})

		DescribeTable("version",
			func(v string) {
				c := &Client{
					Version:                  v,
					Address:                  "127.0.0.1:0",
					User:                     "root",
					Password:                 "nebula",
					ConcurrencyPerAddress:    10,
					ReconnectInitialInterval: time.Second,
					Retry:                    3,
					RetryInitialInterval:     time.Second,
				}

				pool, err := c.BuildClientPool()
				var isSupportVersion = true
				switch v {
				case "":
					v = ClientVersionDefault
				case ClientVersion3:
				default:
					isSupportVersion = false
				}
				if isSupportVersion {
					Expect(c.Version).To(Equal(v))
					Expect(err).NotTo(HaveOccurred())
					Expect(pool).NotTo(BeNil())
				} else {
					Expect(stderrors.Is(err, errors.ErrUnsupportedClientVersion)).To(BeTrue())
					Expect(pool).To(BeNil())
				}
			},
			EntryDescription("%[1]s"),
			Entry(nil, ""),
			Entry(nil, "v3"),
			Entry(nil, "v"),
		)
	})

	DescribeTable("OptimizeFiles",
		func(configPath string, files, expectFiles []string) {
			l := &Log{
				Files: files,
			}
			Expect(l.OptimizeFiles(configPath)).NotTo(HaveOccurred())
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
