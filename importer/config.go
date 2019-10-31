package importer

import (
	"io/ioutil"
	"log"

	yaml "gopkg.in/yaml.v2"
)

type Settings struct {
	Retry       int `yaml:"retry"`
	Concurrency int `yaml:"concurrency"`
	Connection  struct {
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		Address  string `yaml:"address"`
	} `yaml:"connection"`
}

type Prop struct {
	Name string `yaml:"name"`
	Type string `yaml:"type"`
}

type Desc struct {
	Name  string `yaml:"name"`
	Props []Prop `yaml:"props"`
}

type Edge Desc
type Tag Desc

type Vertex struct {
	Tags []Tag `yaml:"tags"`
}

type Schema struct {
	Space  string `yaml:"space"`
	Type   string `yaml:"type"`
	Edge   Edge   `yaml:"edge"`
	Vertex Vertex `yaml:"vertex"`
}

type ErrConfig struct {
	FailDataPath string `yaml:"failDataPath"`
	LogPath      string `yaml:"logPath"`
}

type CSVConfig struct {
	WithHeader bool `yaml:"withHeader"`
}

type File struct {
	Path      string    `yaml:"path"`
	BatchSize int       `yaml:"batchSize"`
	Type      string    `yaml:"type"`
	CSV       CSVConfig `yaml:"csv"`
	Schema    Schema    `yaml:"schema"`
	Error     ErrConfig `yaml:"error"`
}

type YAMLConfig struct {
	Version     string   `yaml:"version"`
	Description string   `yaml:"description"`
	Settings    Settings `yaml:"settings"`
	Files       []File   `yaml:"files"`
}

func Parse(filename string) (YAMLConfig, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return YAMLConfig{}, err
	}

	var conf YAMLConfig
	if err = yaml.Unmarshal(content, &conf); err != nil {
		log.Fatal(err)
	}

	return conf, nil
}
