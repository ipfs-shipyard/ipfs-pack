package main

import (
	"strconv"
	"unicode/utf8"
)

// unescape unescapes a strict subset of c-style escapes sequences:
// \' or \" not allowed; octals (\034) must always be 3 digits; hexs
// (\xAF) must always be 2 digits
func unescape(str string) (string, error) {
	buf := make([]byte, len(str))
	offset := 0
	for len(str) > 0 {
		val, _, tail, err := strconv.UnquoteChar(str, 0)
		if err != nil {
			return "", err
		}
		offset += utf8.EncodeRune(buf[offset:], val)
		str = tail
	}
	return string(buf[0:offset]), nil
}

// escape escapes problematic characters using c-style backslashes.
// Currently escapes: newline, c.r., tab, and backslash
func escape(str string) string {
	buf := make([]byte, 0, len(str)*2)
	for i := 0; i < len(str); i++ {
		switch str[i] {
		case '\n':
			buf = append(buf, '\\', 'n')
		case '\r':
			buf = append(buf, '\\', 'r')
		case '\t':
			buf = append(buf, '\\', 't')
		case '\\':
			buf = append(buf, '\\', '\\')
		default:
			buf = append(buf, str[i])
		}
	}
	return string(buf)
}
