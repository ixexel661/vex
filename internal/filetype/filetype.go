package filetype

import (
	"fmt"

	"github.com/ixexel661/vex/internal/highlight"
)

// Type describes how vex handles one kind of file. Every capability is
// optional, nil means the file type does not support it.
type Type struct {
	Name        string
	Extensions  []string
	Highlighter highlight.Highlighter
	Formatter   Formatter
	Validator   Validator
}

type Formatter interface {
	Format(src, indent string) (string, error)
}

type Validator interface {
	Validate(src string) error
}

// SyntaxError is a problem at a line of a document, counted from 1.
type SyntaxError struct {
	Line    int
	Message string
}

func (e *SyntaxError) Error() string {
	return fmt.Sprintf("line %d: %s", e.Line, e.Message)
}

// Plain is used for files no registered type claims.
func Plain() Type {
	return Type{Name: "Text"}
}
