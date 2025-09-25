package config

import (
	"bytes"
	stderrors "errors"

	configbase "github.com/vesoft-inc/nebula-importer/v4/pkg/config/base"
	configv3 "github.com/vesoft-inc/nebula-importer/v4/pkg/config/v3"
	"github.com/vesoft-inc/nebula-importer/v4/pkg/errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v3"
)

var _ = Describe("FromFile", func() {
	It("successfully v3", func() {
		c1, err := FromFile("testdata/nebula-importer.v3.yaml")
		Expect(err).NotTo(HaveOccurred())

		cv3, ok := c1.(*configv3.Config)
		Expect(ok).To(BeTrue())
		Expect(cv3).NotTo(BeNil())

		Expect(cv3.Client.Version).To(Equal(configbase.ClientVersion3))
		Expect(cv3.Manager.GraphName).To(Equal("graphName"))
		Expect(cv3.Manager.GraphName).To(Equal("graphName"))
		Expect(cv3.Sources).To(HaveLen(3))
		Expect(cv3.Sources[0].Nodes).To(HaveLen(2))
		Expect(cv3.Sources[0].Edges).To(HaveLen(0))
		Expect(cv3.Sources[1].Nodes).To(HaveLen(0))
		Expect(cv3.Sources[1].Edges).To(HaveLen(2))
		Expect(cv3.Sources[2].Nodes).To(HaveLen(2))
		Expect(cv3.Sources[2].Edges).To(HaveLen(2))

		content, err := yaml.Marshal(c1)
		Expect(err).NotTo(HaveOccurred())
		Expect(content).NotTo(BeEmpty())

		c2, err := FromBytes(content)
		Expect(err).NotTo(HaveOccurred())
		Expect(c2).To(Equal(c1))

		c3, err := FromReader(bytes.NewReader(content))
		Expect(err).NotTo(HaveOccurred())
		Expect(c3).To(Equal(c1))
	})

	It("json configuration", func() {
		c1, err := FromFile("testdata/nebula-importer.v3.yaml")
		Expect(err).NotTo(HaveOccurred())

		c2, err := FromFile("testdata/nebula-importer.v3.json")
		Expect(err).NotTo(HaveOccurred())

		Expect(c2).To(Equal(c1))
	})

	It("configuration file not exists", func() {
		c, err := FromFile("testdata/not-exists.yaml")
		Expect(err).To(HaveOccurred())
		Expect(c).To(BeNil())
	})
})

var _ = Describe("FromBytes", func() {
	It("Unmarshal failed 1", func() {
		c, err := FromBytes([]byte(`
client:
  version: : v

`))
		Expect(err).To(HaveOccurred())
		Expect(c).To(BeNil())
	})

	It("Unmarshal failed 2", func() {
		c, err := FromBytes([]byte(`
client:
  version: v3
log:
  files: ""
`))
		Expect(err).To(HaveOccurred())
		Expect(c).To(BeNil())
	})

	It("unsupported client version failed", func() {
		c, err := FromBytes([]byte(`
client:
  version: v
`))
		Expect(err).To(HaveOccurred())
		Expect(stderrors.Is(err, errors.ErrUnsupportedClientVersion)).To(BeTrue())
		Expect(c).To(BeNil())
	})
})

type testErrorReader struct {
	err error
}

func (r testErrorReader) Read([]byte) (n int, err error) {
	return 0, r.err
}

var _ = Describe("FromBytes", func() {
	It("Unmarshal failed 1", func() {
		c, err := FromReader(testErrorReader{
			err: stderrors.New("read failed"),
		})
		Expect(err).To(HaveOccurred())
		Expect(c).To(BeNil())
	})
})
