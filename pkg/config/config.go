package config

import (
	"encoding/json"
	"io"
	"os"

	configbase "github.com/vesoft-inc/nebula-importer/v4/pkg/config/base"
	configv3 "github.com/vesoft-inc/nebula-importer/v4/pkg/config/v3"
	"github.com/vesoft-inc/nebula-importer/v4/pkg/errors"

	"gopkg.in/yaml.v3"
)

type (
	Client       = configbase.Client
	Log          = configbase.Log
	Configurator = configbase.Configurator
)

func FromBytes(content []byte) (Configurator, error) {
	if newContent, err := jsonToYaml(content); err == nil {
		content = newContent
	}

	type tmpConfig struct {
		Client struct {
			Version string `yaml:"version"`
		} `yaml:"client"`
	}
	var tc tmpConfig
	if err := yaml.Unmarshal(content, &tc); err != nil {
		return nil, err
	}
	var c Configurator
	switch tc.Client.Version {
	case configbase.ClientVersion3:
		c = &configv3.Config{}
	default:
		return nil, errors.ErrUnsupportedClientVersion
	}

	if err := yaml.Unmarshal(content, c); err != nil {
		return nil, err
	}
	return c, nil
}

func FromReader(r io.Reader) (Configurator, error) {
	content, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return FromBytes(content)
}

func FromFile(name string) (Configurator, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return FromReader(f)
}

func jsonToYaml(content []byte) ([]byte, error) {
	var jsonObj any
	err := json.Unmarshal(content, &jsonObj)
	if err != nil {
		return nil, err
	}

	return yaml.Marshal(jsonObj)
}
