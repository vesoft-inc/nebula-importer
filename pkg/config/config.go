package config

import (
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

type NebulaClientConnection struct {
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Address  string `yaml:"address"`
}

type NebulaClientSettings struct {
	Concurrency int                    `yaml:"concurrency"`
	Connection  NebulaClientConnection `yaml:"connection"`
}

type Prop struct {
	Name string `yaml:"name"`
	Type string `yaml:"type"`
}

type Edge struct {
	Name        string `yaml:"name"`
	WithRanking bool   `yaml:"withRanking"`
	Props       []Prop `yaml:"props"`
}

type Tag struct {
	Name  string `yaml:"name"`
	Props []Prop `yaml:"props"`
}

type Vertex struct {
	Tags []Tag `yaml:"tags"`
}

type Schema struct {
	Space  string `yaml:"space"`
	Type   string `yaml:"type"`
	Edge   Edge   `yaml:"edge"`
	Vertex Vertex `yaml:"vertex"`
}

type CSVConfig struct {
	WithHeader bool `yaml:"withHeader"`
	WithLabel  bool `yaml:"withLabel"`
}

type File struct {
	Path         string    `yaml:"path"`
	FailDataPath string    `yaml:"failDataPath"`
	BatchSize    int       `yaml:"batchSize"`
	Type         string    `yaml:"type"`
	CSV          CSVConfig `yaml:"csv"`
	Schema       Schema    `yaml:"schema"`
}

type YAMLConfig struct {
	Version              string               `yaml:"version"`
	Description          string               `yaml:"description"`
	NebulaClientSettings NebulaClientSettings `yaml:"clientSettings"`
	LogPath              string               `yaml:"logPath"`
	Files                []File               `yaml:"files"`
}

func Parse(filename string) (*YAMLConfig, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var conf YAMLConfig
	if err = yaml.Unmarshal(content, &conf); err != nil {
		return nil, err
	}

	return &conf, nil
}
