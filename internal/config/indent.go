package config

import (
	"fmt"
	"strconv"
	"strings"
)

// Indent is the whitespace of one indentation level. The config file writes
// it as "tabs" or as a number of spaces, e.g. "2 spaces".
type Indent string

const maxIndentSpaces = 8

func (i Indent) String() string {
	switch {
	case i == "\t":
		return "tabs"
	case len(i) == 1:
		return "1 space"
	default:
		return fmt.Sprintf("%d spaces", len(i))
	}
}

func (i Indent) MarshalText() ([]byte, error) {
	return []byte(i.String()), nil
}

func (i *Indent) UnmarshalText(text []byte) error {
	s := string(text)
	if s == "tabs" {
		*i = "\t"
		return nil
	}
	count, unit, _ := strings.Cut(s, " ")
	n, err := strconv.Atoi(count)
	if err != nil || n < 1 || n > maxIndentSpaces || unit != "space" && unit != "spaces" {
		return fmt.Errorf(`indent must be "tabs" or 1 to %d spaces like "2 spaces", not %q`, maxIndentSpaces, s)
	}
	*i = Indent(strings.Repeat(" ", n))
	return nil
}
