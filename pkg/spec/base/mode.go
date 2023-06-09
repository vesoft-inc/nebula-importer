package specbase

import "strings"

const (
	DefaultMode      = InsertMode
	InsertMode  Mode = "INSERT"
	UpdateMode  Mode = "UPDATE"
	DeleteMode  Mode = "DELETE"
)

type Mode string

func (m Mode) Convert() Mode {
	if m == "" {
		return DefaultMode
	}
	return Mode(strings.ToUpper(string(m)))
}

func (m Mode) IsSupport() bool {
	return m == InsertMode || m == UpdateMode || m == DeleteMode
}
