package importer

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestImporter(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Pkg importer Suite")
}
