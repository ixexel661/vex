package filetype

import (
	"fmt"
	"path/filepath"

	"github.com/ixexel661/vex/internal/buffer"
	"github.com/ixexel661/vex/internal/config"
	"github.com/ixexel661/vex/internal/highlight"
)

// Type describes how vex handles one kind of file. Every capability is
// optional, nil means the file type does not support it.
type Type struct {
	Name       string
	Extensions []string
	// Template is the text a new file starts with.
	Template    string
	Highlighter highlight.Highlighter
	Formatter   Formatter
	Validator   Validator
	Tags        Tags
}

// FileName returns name with the first extension of the type if it has none.
func (t Type) FileName(name string) string {
	if filepath.Ext(name) != "" || len(t.Extensions) == 0 {
		return name
	}
	return name + t.Extensions[0]
}

type Formatter interface {
	Format(src string, opts config.Format) (string, error)
}

// Format formats src with the formatter of the file type.
func (t Type) Format(src string, opts config.Format) (string, error) {
	if t.Formatter == nil {
		return "", fmt.Errorf("%s files cannot be formatted", t.Name)
	}
	formatted, err := t.Formatter.Format(src, opts)
	if err != nil {
		return "", fmt.Errorf("cannot format: %w", err)
	}
	return formatted, nil
}

type Validator interface {
	Validate(src string) error
}

// Tags pairs the start and end tags of elements, so the editor can mark and
// rename both names at once and add the end tag as a start tag is typed.
type Tags interface {
	// Pair returns the name of the tag at p and the name of the other tag of its element.
	Pair(src Source, p buffer.Position) (name, other buffer.Range, ok bool)
	// IsName reports whether text can go into a tag name.
	IsName(text string) bool
	// Close returns the end tag for the start tag that the ">" right before p
	// ends, unless the element has one already.
	Close(src Source, p buffer.Position) (string, bool)
}

// Source is a document as file types read it: its lines and the state the
// highlighter is in at the start of each line.
type Source interface {
	LineCount() int
	Line(row int) string
	State(row int) highlight.State
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
