package reader

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestReader(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Pkg reader Suite")
}
