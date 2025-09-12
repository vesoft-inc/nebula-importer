package util

import (
	"fmt"
	"os"
	"strings"
)

var (
	fnFprint = fmt.Fprint
	fnExit   = os.Exit
)

func CheckErr(err error) {
	switch err.(type) {
	case nil:
		return
	default:
		fatal(fmt.Sprintf("%+v", err), 1)
	}
}

func fatal(msg string, code int) {
	if msg != "" {
		if !strings.HasSuffix(msg, "\n") {
			msg += "\n"
		}
		_, _ = fnFprint(os.Stderr, msg)
	}
	fnExit(code) //revive:disable-line:deep-exit
}
