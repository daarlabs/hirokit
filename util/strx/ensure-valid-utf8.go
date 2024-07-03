package strx

import "unicode/utf8"

func EnsureValidUTF8(s string) string {
	if utf8.ValidString(s) {
		return s
	}
	valid := make([]rune, 0, len(s))
	for i, r := range s {
		if r == utf8.RuneError {
			_, size := utf8.DecodeRuneInString(s[i:])
			if size == 1 {
				continue
			}
		}
		valid = append(valid, r)
	}
	return string(valid)
}
