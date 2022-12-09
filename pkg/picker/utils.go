package picker

func isUnsignedInteger(s string) bool {
	switch len(s) {
	case 0:
		return false
	case 1:
		return isDigit(s[0])
	case 2:
		return isDigit(s[0]) && isDigit(s[1])
	}
	return isIntegerSlow(s)
}

func isIntegerSlow(s string) bool {
	f := isDigit
	if len(s) > 2 && s[0] == '0' && (s[1] == 'x' || s[1] == 'X') {
		s = s[2:]
		f = isHexDigit
	}

	for _, b := range []byte(s) {
		if !f(b) {
			return false
		}
	}
	return true
}

func isDigit(b byte) bool {
	return '0' <= b && b <= '9'
}

func isHexDigit(b byte) bool {
	return isDigit(b) || ('a' <= b && b <= 'f') || ('A' <= b && b <= 'F')
}
