package client

import (
	"log/slog"

	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("nebulaLogger", func() {
	It("newNebulaLogger", func() {
		l := newNebulaLogger(slog.Default())
		l.Info("")
		l.Warn("")
		l.Error("")
	})
})
