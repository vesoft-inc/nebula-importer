package nebula_importer

import (
	"log"
	"testing"
)

func TestYAMLParser(t *testing.T) {
	yaml, err := Parse("../example/example.yaml")
	if err != nil {
		t.Error(err)
	}

	if len(yaml.Files) > 0 {
		log.Printf("num files: %v", yaml.Files[0].Path)
	} else {
		t.Fatal("parse error")
	}
}
