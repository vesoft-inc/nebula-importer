package nebula_importer

import (
	"io/ioutil"
	"log"

	yaml "gopkg.in/yaml.v2"
)

type YAMLConfig struct {
	Version     string `yaml:"version"`
	Description string `yaml:"description"`
	Settings    struct {
		Retry       int `yaml:"retry"`
		Concurrency int `yaml:"concurrency"`
		Connection  struct {
			User     string `yaml:"user"`
			Password string `yaml:"password"`
			Address  string `yaml:"address"`
		} `yaml:"connection"`
	} `yaml:"settings"`
	Files []struct {
		Path   string `yaml:"path"`
		Type   string `yaml:"type"`
		Schema struct {
			Space string `yaml:"space"`
			Name  string `yaml:"name"`
			Type  string `yaml:"type"`
			Props []struct {
				Name string `yaml:"name"`
				Type string `yaml:"type"`
			} `yaml:"props"`
		} `yaml:"schema"`
		Error struct {
			FailDataPath string `yaml:"failDataPath"`
			LogPath      string `yaml:"logPath"`
		} `yaml:"error"`
	} `yaml:"files"`
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
