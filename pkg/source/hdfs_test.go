//go:build linux

package source

import (
	stderrors "errors"
	"io"
	"os"
	"testing/fstest"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/colinmarc/hdfs/v2"
	"github.com/colinmarc/hdfs/v2/hadoopconf"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("hdfsSource", func() {
	var (
		address        = "nn1:9000,nn2:9000"
		user           = "user"
		content        = []byte("Hello")
		patches        *gomonkey.Patches
		hdfsClient     = &hdfs.Client{}
		hdfsFileReader = &hdfs.FileReader{}
	)
	BeforeEach(func() {
		patches = gomonkey.NewPatches()
		mockFile, err := fstest.MapFS{
			"file": {
				Data: content,
			},
		}.Open("file")
		Expect(err).NotTo(HaveOccurred())

		patches.ApplyFunc(hdfs.NewClient, func(opts hdfs.ClientOptions) (*hdfs.Client, error) {
			Expect(opts.Addresses).To(Equal([]string{"nn1:9000", "nn2:9000"}))
			Expect(opts.User).To(Equal(user))
			return hdfsClient, nil
		})
		patches.ApplyMethodReturn(hdfsClient, "Open", hdfsFileReader, nil)
		patches.ApplyMethodReturn(hdfsClient, "Close", nil)

		patches.ApplyMethod(hdfsFileReader, "Stat", func() os.FileInfo {
			fi, err := mockFile.Stat()
			Expect(err).NotTo(HaveOccurred())
			return fi
		})
		patches.ApplyMethod(hdfsFileReader, "Read", func(_ *hdfs.FileReader, p []byte) (int, error) {
			return mockFile.Read(p)
		})
		patches.ApplyMethod(hdfsFileReader, "Close", func(_ *hdfs.FileReader) error {
			return mockFile.Close()
		})
	})
	AfterEach(func() {
		patches.Reset()
	})
	It("successfully", func() {
		c := Config{
			HDFS: &HDFSConfig{
				Address: address,
				User:    user,
				Path:    "file",
			},
		}

		s, err := New(&c)
		Expect(err).NotTo(HaveOccurred())
		Expect(s).To(BeAssignableToTypeOf(&hdfsSource{}))

		Expect(s.Name()).To(Equal("hdfs nn1:9000,nn2:9000 file"))

		Expect(s.Config()).NotTo(BeNil())

		err = s.Open()
		Expect(err).NotTo(HaveOccurred())

		sz, err := s.Size()
		Expect(err).NotTo(HaveOccurred())
		Expect(sz).To(Equal(int64(len(content))))

		var p [32]byte
		n, err := s.Read(p[:])
		Expect(err).NotTo(HaveOccurred())
		Expect(n).To(Equal(len(content)))
		Expect(p[:n]).To(Equal(content))

		for i := 0; i < 2; i++ {
			n, err = s.Read(p[:])
			Expect(err).To(Equal(io.EOF))
			Expect(n).To(Equal(0))
		}

		err = s.Close()
		Expect(err).NotTo(HaveOccurred())
	})

	It("LoadFromEnvironment failed", func() {
		c := Config{
			HDFS: &HDFSConfig{
				Address: address,
				User:    user,
				Path:    "file",
			},
		}

		patches.ApplyFuncReturn(hadoopconf.LoadFromEnvironment, hadoopconf.HadoopConf(nil), stderrors.New("test error"))

		s, err := New(&c)
		Expect(err).NotTo(HaveOccurred())
		Expect(s).To(BeAssignableToTypeOf(&hdfsSource{}))

		err = s.Open()
		Expect(err).To(HaveOccurred())
	})

	It("hdfs.NewClient failed", func() {
		c := Config{
			HDFS: &HDFSConfig{
				Address: address,
				User:    user,
				Path:    "file",
			},
		}

		patches.ApplyFuncReturn(hdfs.NewClient, nil, stderrors.New("test error"))

		s, err := New(&c)
		Expect(err).NotTo(HaveOccurred())
		Expect(s).To(BeAssignableToTypeOf(&hdfsSource{}))

		err = s.Open()
		Expect(err).To(HaveOccurred())
	})

	It("Open failed", func() {
		c := Config{
			HDFS: &HDFSConfig{
				Address: address,
				User:    user,
				Path:    "file",
			},
		}

		patches.ApplyMethodReturn(hdfsClient, "Open", nil, stderrors.New("test error"))

		s, err := New(&c)
		Expect(err).NotTo(HaveOccurred())
		Expect(s).To(BeAssignableToTypeOf(&hdfsSource{}))

		err = s.Open()
		Expect(err).To(HaveOccurred())
	})
})
