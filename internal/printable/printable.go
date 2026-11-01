// Package printable makes text safe to show in a terminal.
package printable

import (
	"strings"
	"unicode"
)

// String replaces control characters with "?", so text from files cannot
// send escape sequences to the terminal. A tab becomes a space.
func String(s string) string {
	return strings.Map(func(r rune) rune {
		switch {
		case r == '\t':
			return ' '
		case unicode.IsControl(r):
			return '?'
		default:
			return r
		}
	}, s)
}
