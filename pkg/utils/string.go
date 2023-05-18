package utils

import "strings"

func IsInteger(s string) bool {
	if s == "" {
		return false
	}
	if s[0] == '+' || s[0] == '-' {
		s = s[1:]
	}
	return IsUnsignedInteger(s)
}

func IsUnsignedInteger(s string) bool {
	switch len(s) {
	case 0:
		return false
	case 1:
		return IsDigit(s[0])
	case 2:
		return IsDigit(s[0]) && IsDigit(s[1])
	}
	return isUnsignedIntegerSlow(s)
}

func isUnsignedIntegerSlow(s string) bool {
	f := IsDigit
	if len(s) > 2 && s[0] == '0' && (s[1] == 'x' || s[1] == 'X') {
		s = s[2:]
		f = IsHexDigit
	}

	for _, b := range []byte(s) {
		if !f(b) {
			return false
		}
	}
	return true
}

func IsDigit(b byte) bool {
	return '0' <= b && b <= '9'
}

func IsHexDigit(b byte) bool {
	return IsDigit(b) || ('a' <= b && b <= 'f') || ('A' <= b && b <= 'F')
}

func ConvertIdentifier(s string) string {
	const (
		backslashChar   = '\\'
		backslashString = string(backslashChar)
		backtickChar    = '`'
		backtickString  = string(backtickChar)
	)
	if strings.IndexByte(s, backslashChar) >= 0 {
		s = strings.ReplaceAll(s, backslashString, backslashString+backslashString)
	}
	if strings.IndexByte(s, backtickChar) >= 0 {
		s = strings.ReplaceAll(s, backtickString, backslashString+backtickString)
	}
	return backtickString + s + backtickString
}
