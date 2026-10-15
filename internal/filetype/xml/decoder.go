package xml

import (
	stdxml "encoding/xml"
	"errors"
	"io"
	"strings"

	"github.com/ixexel661/vex/internal/filetype"
)

func newDecoder(src string) *stdxml.Decoder {
	d := stdxml.NewDecoder(strings.NewReader(src))
	// The text is UTF-8 already, whatever encoding the XML declaration names.
	d.CharsetReader = func(_ string, input io.Reader) (io.Reader, error) {
		return input, nil
	}
	return d
}

func toSyntaxError(err error) error {
	var syntaxErr *stdxml.SyntaxError
	if errors.As(err, &syntaxErr) {
		return &filetype.SyntaxError{Line: syntaxErr.Line, Message: syntaxErr.Msg}
	}
	return err
}
