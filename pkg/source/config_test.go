package source

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config", func() {
	Describe(".Clone", func() {
		It("S3", func() {
			c := Config{
				S3: &S3Config{
					Key: "key",
				},
			}
			c1 := c.Clone()
			Expect(c1.S3.Key).To(Equal("key"))
			c.S3.Key = "x"
			Expect(c1.S3.Key).To(Equal("key"))
		})

		It("OSS", func() {
			c := Config{
				OSS: &OSSConfig{
					Key: "key",
				},
			}
			c1 := c.Clone()
			Expect(c1.OSS.Key).To(Equal("key"))
			c.OSS.Key = "x"
			Expect(c1.OSS.Key).To(Equal("key"))
		})

		It("FTP", func() {
			c := Config{
				FTP: &FTPConfig{
					Path: "path",
				},
			}
			c1 := c.Clone()
			Expect(c1.FTP.Path).To(Equal("path"))
			c.FTP.Path = "x"
			Expect(c1.FTP.Path).To(Equal("path"))
		})

		It("SFTP", func() {
			c := Config{
				SFTP: &SFTPConfig{
					Path: "path",
				},
			}
			c1 := c.Clone()
			Expect(c1.SFTP.Path).To(Equal("path"))
			c.SFTP.Path = "x"
			Expect(c1.SFTP.Path).To(Equal("path"))
		})

		It("HDFS", func() {
			c := Config{
				HDFS: &HDFSConfig{
					Path: "path",
				},
			}
			c1 := c.Clone()
			Expect(c1.HDFS.Path).To(Equal("path"))
			c.HDFS.Path = "x"
			Expect(c1.HDFS.Path).To(Equal("path"))
		})

		It("Local", func() {
			c := Config{
				Local: &LocalConfig{
					Path: "path",
				},
			}
			c1 := c.Clone()
			Expect(c1.Local.Path).To(Equal("path"))
			c.Local.Path = "x"
			Expect(c1.Local.Path).To(Equal("path"))
		})
	})
})
