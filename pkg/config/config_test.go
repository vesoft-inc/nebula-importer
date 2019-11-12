package config

import (
	"fmt"
	"strings"
	"testing"
)

func TestYAMLParser(t *testing.T) {
	yaml, err := Parse("../../example/example.yaml")
	if err != nil {
		t.Fatal(err)
	}

	for i, file := range yaml.Files {
		if strings.ToLower(file.Type) != "csv" {
			t.Fatal("Error file type")
		}
		switch strings.ToLower(file.Schema.Type) {
		case "edge":
			if i != 0 {
				t.Fatal("First file is not edge data")
			}
		case "vertex":
			if i != 1 {
				t.Fatal("Second file is not vertex data")
			}
			for j, tag := range file.Schema.Vertex.Tags {
				tagName := fmt.Sprintf("tag%d", j+1)
				if tag.Name != tagName {
					t.Fatalf("Wrong tag name: %s vs. %s", tag.Name, tagName)
				}
			}
		default:
			t.Fatal("Error schema type")
		}
	}
}
