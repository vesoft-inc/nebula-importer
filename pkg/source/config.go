package source

type (
	Config struct {
		Local *LocalConfig `yaml:",inline"`
		S3    *S3Config    `yaml:"s3,omitempty"`
		OSS   *OSSConfig   `yaml:"oss,omitempty"`
		FTP   *FTPConfig   `yaml:"ftp,omitempty"`
		SFTP  *SFTPConfig  `yaml:"sftp,omitempty"`
		HDFS  *HDFSConfig  `yaml:"hdfs,omitempty"`
		GCS   *GCSConfig   `yaml:"gcs,omitempty"`
		// The following is format information
		CSV *CSVConfig `yaml:"csv,omitempty"`
	}

	CSVConfig struct {
		Delimiter  string `yaml:"delimiter,omitempty"`
		Comment    string `yaml:"comment,omitempty"`
		WithHeader bool   `yaml:"withHeader,omitempty"`
		LazyQuotes bool   `yaml:"lazyQuotes,omitempty"`
	}
)

func (c *Config) Clone() *Config {
	cpy := *c
	switch {
	case cpy.S3 != nil:
		cpy1 := *cpy.S3
		cpy.S3 = &cpy1
	case cpy.OSS != nil:
		cpy1 := *cpy.OSS
		cpy.OSS = &cpy1
	case cpy.FTP != nil:
		cpy1 := *cpy.FTP
		cpy.FTP = &cpy1
	case cpy.SFTP != nil:
		cpy1 := *cpy.SFTP
		cpy.SFTP = &cpy1
	case cpy.HDFS != nil:
		cpy1 := *cpy.HDFS
		cpy.HDFS = &cpy1
	case cpy.GCS != nil:
		cpy1 := *cpy.GCS
		cpy.GCS = &cpy1
	default:
		cpy1 := *cpy.Local
		cpy.Local = &cpy1
	}
	return &cpy
}
