package client

import (
	"github.com/vesoft-inc/nebula-importer/v4/pkg/logger"

	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("nebulaLogger", func() {
	It("newNebulaLogger", func() {
		l := newNebulaLogger(logger.NopLogger)
		l.Info("")
		l.Warn("")
		l.Error("")
		l.Fatal("")
	})
})
