package search

import (
	"strings"
	"unicode"
)

// prepare returns the runes to look for. Smart case: the search ignores case
// unless the query contains an upper case letter.
func prepare(query string) (pattern []rune, ignoreCase bool) {
	ignoreCase = !strings.ContainsFunc(query, unicode.IsUpper)
	return normalize(query, ignoreCase), ignoreCase
}

// normalize returns the runes of s, in lower case if case is ignored. Each
// rune is mapped on its own, so positions in the result are positions in s.
func normalize(s string, ignoreCase bool) []rune {
	runes := []rune(s)
	if ignoreCase {
		for i, r := range runes {
			runes[i] = unicode.ToLower(r)
		}
	}
	return runes
}
