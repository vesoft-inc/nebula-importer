package picker

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestPicker(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Pkg picker Suite")
}
