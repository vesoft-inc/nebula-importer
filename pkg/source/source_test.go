package source

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Source", func() {
	It("S3", func() {
		c := Config{
			S3: &S3Config{
				Key: "key",
			},
		}
		s, err := New(&c)
		Expect(err).NotTo(HaveOccurred())
		Expect(s).To(BeAssignableToTypeOf(&s3Source{}))
	})

	It("OSS", func() {
		c := Config{
			OSS: &OSSConfig{
				Key: "key",
			},
		}
		s, err := New(&c)
		Expect(err).NotTo(HaveOccurred())
		Expect(s).To(BeAssignableToTypeOf(&ossSource{}))
	})

	It("FTP", func() {
		c := Config{
			FTP: &FTPConfig{
				Path: "path",
			},
		}
		s, err := New(&c)
		Expect(err).NotTo(HaveOccurred())
		Expect(s).To(BeAssignableToTypeOf(&ftpSource{}))
	})

	It("SFTP", func() {
		c := Config{
			SFTP: &SFTPConfig{
				Path: "path",
			},
		}
		s, err := New(&c)
		Expect(err).NotTo(HaveOccurred())
		Expect(s).To(BeAssignableToTypeOf(&sftpSource{}))
	})

	It("HDFS", func() {
		c := Config{
			HDFS: &HDFSConfig{
				Path: "path",
			},
		}
		s, err := New(&c)
		Expect(err).NotTo(HaveOccurred())
		Expect(s).To(BeAssignableToTypeOf(&hdfsSource{}))
	})

	It("Local", func() {
		c := Config{
			Local: &LocalConfig{
				Path: "path",
			},
		}
		s, err := New(&c)
		Expect(err).NotTo(HaveOccurred())
		Expect(s).To(BeAssignableToTypeOf(&localSource{}))
	})
})
