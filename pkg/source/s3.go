package source

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
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
	ctx := context.Background()
	var cfg aws.Config
	var err error

	optFns := []func(*config.LoadOptions) error{
		config.WithRegion(s.c.S3.Region),
	}

	if s.c.S3.Endpoint != "" {
		customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...any) (aws.Endpoint, error) {
			return aws.Endpoint{
				URL:               s.c.S3.Endpoint,
				HostnameImmutable: true,
			}, nil
		})
		optFns = append(optFns, config.WithEndpointResolverWithOptions(customResolver))
	}

	if s.c.S3.AccessKeyID != "" || s.c.S3.AccessKeySecret != "" || s.c.S3.Token != "" {
		optFns = append(optFns, config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(s.c.S3.AccessKeyID, s.c.S3.AccessKeySecret, s.c.S3.Token),
		))
	}

	cfg, err = config.LoadDefaultConfig(ctx, optFns...)
	if err != nil {
		return err
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	obj, err := client.GetObject(ctx, &s3.GetObjectInput{
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
