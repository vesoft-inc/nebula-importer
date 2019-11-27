package config

import (
	"strings"
	"testing"
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
