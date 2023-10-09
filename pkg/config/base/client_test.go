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

	It(".OptimizePath nil", func() {
		var c *Client
		Expect(c.OptimizePath("")).NotTo(HaveOccurred())
	})

	DescribeTable("OptimizePath ssl",
		func(configPath string, ssl, expectSSL *SSL) {
			c := Client{
				SSL: ssl,
			}
			Expect(c.OptimizePath(configPath)).NotTo(HaveOccurred())
			Expect(c.SSL).To(Equal(expectSSL))
		},
		EntryDescription("%[1]s : %[2]v => %[3]v"),

		Entry(nil, "f.yaml", nil, nil),
		Entry(nil, "f.yaml", &SSL{}, &SSL{}),
		Entry(nil, "./f.yaml",
			&SSL{
				Enable:   true,
				CertPath: "cert.crt",
				KeyPath:  "d10/cert.key",
				CAPath:   "d20/ca.crt",
			},
			&SSL{
				Enable:   true,
				CertPath: "cert.crt",
				KeyPath:  "d10/cert.key",
				CAPath:   "d20/ca.crt",
			},
		),
		Entry(nil, "./d1/f.yaml",
			&SSL{
				Enable:   true,
				CertPath: "cert.crt",
				KeyPath:  "d10/cert.key",
				CAPath:   "d20/ca.crt",
			},
			&SSL{
				Enable:   true,
				CertPath: "d1/cert.crt",
				KeyPath:  "d1/d10/cert.key",
				CAPath:   "d1/d20/ca.crt",
			},
		),
		Entry(nil, "./d1/f.yaml",
			&SSL{
				Enable:   true,
				CertPath: "/cert.crt",
				KeyPath:  "/d10/cert.key",
				CAPath:   "/d20/ca.crt",
			},
			&SSL{
				Enable:   true,
				CertPath: "/cert.crt",
				KeyPath:  "/d10/cert.key",
				CAPath:   "/d20/ca.crt",
			},
		),
	)
})
