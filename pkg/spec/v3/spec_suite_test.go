package specv3

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSpec(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Pkg spec v3 Suite")
}
