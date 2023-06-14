package source

import (
	"context"
	"fmt"
	"os"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

var _ Source = (*gcsSource)(nil)

type (
	GCSConfig struct {
		Endpoint        string `yaml:"endpoint,omitempty"`
		CredentialsFile string `yaml:"credentialsFile,omitempty"`
		CredentialsJSON string `yaml:"credentialsJSON,omitempty"`
		Bucket          string `yaml:"bucket,omitempty"`
		Key             string `yaml:"key,omitempty"`
	}

	gcsSource struct {
		c      *Config
		reader *storage.Reader
	}
)

func newGCSSource(c *Config) Source {
	return &gcsSource{
		c: c,
	}
}

func (s *gcsSource) Name() string {
	return s.c.GCS.String()
}

func (s *gcsSource) Open() error {
	if s.c.GCS.Endpoint != "" {
		err := os.Setenv("STORAGE_EMULATOR_HOST", s.c.GCS.Endpoint)
		if err != nil {
			return err
		}
	}

	var gcsOption option.ClientOption
	if s.c.GCS.CredentialsFile != "" {
		gcsOption = option.WithCredentialsFile(s.c.GCS.CredentialsFile)
	} else if s.c.GCS.CredentialsJSON != "" {
		gcsOption = option.WithCredentialsJSON([]byte(s.c.GCS.CredentialsJSON))
	} else {
		gcsOption = option.WithoutAuthentication()
	}

	ctx := context.Background()
	client, err := storage.NewClient(ctx, gcsOption)
	if err != nil {
		return err
	}
	defer client.Close()

	obj := client.Bucket(s.c.GCS.Bucket).Object(s.c.GCS.Key)
	if s.reader, err = obj.NewReader(ctx); err != nil {
		return err
	}
	return nil
}

func (s *gcsSource) Config() *Config {
	return s.c
}

func (s *gcsSource) Size() (int64, error) {
	return s.reader.Attrs.Size, nil
}

func (s *gcsSource) Read(p []byte) (int, error) {
	return s.reader.Read(p)
}

func (s *gcsSource) Close() error {
	return s.reader.Close()
}

func (c *GCSConfig) String() string {
	return fmt.Sprintf("gcs %s %s/%s", c.Endpoint, c.Bucket, c.Key)
}
