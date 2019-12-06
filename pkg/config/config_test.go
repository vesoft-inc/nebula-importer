package config

import (
	"strings"
	"testing"

	yaml "gopkg.in/yaml.v2"
)

func TestYAMLParser(t *testing.T) {
	yaml, err := Parse("../../examples/example.yaml")
	if err != nil {
		t.Fatal(err)
	}

	for _, file := range yaml.Files {
		if strings.ToLower(file.Type) != "csv" {
			t.Fatal("Error file type")
		}
		switch strings.ToLower(file.Schema.Type) {
		case "edge":
		case "vertex":
			if len(file.Schema.Vertex.Tags) == 0 && !file.CSV.WithHeader {
				t.Fatal("Empty tags in vertex")
			}
		default:
			t.Fatal("Error schema type")
		}
	}
}

type testYAML struct {
	Version *string `yaml:"version"`
	Files   *[]struct {
		Path *string `yaml:"path"`
	} `yaml:"files"`
}

var yamlStr string = `
version: beta
files:
  - path: ./file.csv
`

func TestTypePointer(t *testing.T) {
	ty := testYAML{}
	if err := yaml.Unmarshal([]byte(yamlStr), &ty); err != nil {
		t.Fatal(err)
	}
	t.Logf("yaml: %v, %v", *ty.Version, *ty.Files)
}
