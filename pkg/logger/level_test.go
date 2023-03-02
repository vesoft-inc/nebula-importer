package logger

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Level", func() {
	DescribeTable("cases",
		func(text string, lvl Level) {
			l := ParseLevel(text)
			Expect(l).To(Equal(lvl))
			if text == "" {
				text = "info"
			}
			Expect(l.String()).To(Equal(strings.ToUpper(text)))
		},
		EntryDescription("%[1]s"),
		Entry(nil, "", InfoLevel),
		Entry(nil, "debug", DebugLevel),
		Entry(nil, "info", InfoLevel),
		Entry(nil, "Info", InfoLevel),
		Entry(nil, "INFO", InfoLevel),
		Entry(nil, "warn", WarnLevel),
		Entry(nil, "error", ErrorLevel),
		Entry(nil, "panic", PanicLevel),
		Entry(nil, "fatal", FatalLevel),
	)
})
