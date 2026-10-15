package xml

import (
	stdxml "encoding/xml"
	"errors"
	"io"
	"strings"

	"github.com/ixexel661/vex/internal/filetype"
)

// Validator checks that a document is well-formed.
type Validator struct{}

func (Validator) Validate(src string) error {
	d := newDecoder(src)
	depth, roots := 0, 0
	for {
		tok, err := d.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return toSyntaxError(err)
		}
		switch t := tok.(type) {
		case stdxml.StartElement:
			if depth == 0 {
				roots++
				if roots > 1 {
					return errorAt(d, "more than one root element")
				}
			}
			depth++
		case stdxml.EndElement:
			depth--
		case stdxml.CharData:
			if depth == 0 && strings.TrimSpace(string(t)) != "" {
				return errorAt(d, "text outside the root element")
			}
		}
	}
	if roots == 0 {
		return &filetype.SyntaxError{Line: 1, Message: "no root element"}
	}
	return nil
}

func errorAt(d *stdxml.Decoder, message string) error {
	line, _ := d.InputPos()
	return &filetype.SyntaxError{Line: line, Message: message}
}
