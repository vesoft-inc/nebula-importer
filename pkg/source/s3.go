package source

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

var _ Source = (*s3Source)(nil)

type (
	S3Config struct {
		Endpoint        string `yaml:"endpoint,omitempty"`
		Region          string `yaml:"region,omitempty"`
		AccessKeyID     string `yaml:"accessKeyID,omitempty"`
		AccessKeySecret string `yaml:"accessKeySecret,omitempty"`
		Token           string `yaml:"token,omitempty"`
		Bucket          string `yaml:"bucket,omitempty"`
		Key             string `yaml:"key,omitempty"`
	}

	s3Source struct {
		c   *Config
		obj *s3.GetObjectOutput
	}
)

func newS3Source(c *Config) Source {
	return &s3Source{
		c: c,
	}
}

func (s *s3Source) Name() string {
	return s.c.S3.String()
}

func (s *s3Source) Open() error {
	awsConfig := &aws.Config{
		Region:           aws.String(s.c.S3.Region),
		Endpoint:         aws.String(s.c.S3.Endpoint),
		S3ForcePathStyle: aws.Bool(true),
	}

	if s.c.S3.AccessKeyID != "" || s.c.S3.AccessKeySecret != "" || s.c.S3.Token != "" {
		awsConfig.Credentials = credentials.NewStaticCredentials(s.c.S3.AccessKeyID, s.c.S3.AccessKeySecret, s.c.S3.Token)
	}

	sess, err := session.NewSession(awsConfig)
	if err != nil {
		return err
	}

	svc := s3.New(sess)

	obj, err := svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(s.c.S3.Bucket),
		Key:    aws.String(strings.TrimLeft(s.c.S3.Key, "/")),
	})
	if err != nil {
		return err
	}

	s.obj = obj

	return nil
}

func (s *s3Source) Config() *Config {
	return s.c
}

func (s *s3Source) Size() (int64, error) {
	return *s.obj.ContentLength, nil
}

func (s *s3Source) Read(p []byte) (int, error) {
	return s.obj.Body.Read(p)
}

func (s *s3Source) Close() error {
	return s.obj.Body.Close()
}

func (c *S3Config) String() string {
	return fmt.Sprintf("s3 %s:%s %s/%s", c.Region, c.Endpoint, c.Bucket, c.Key)
}
