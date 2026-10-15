package xml

import (
	"errors"
	"testing"

	"github.com/ixexel661/vex/internal/filetype"
)

func TestValidateAcceptsWellFormedXML(t *testing.T) {
	src := "<?xml version=\"1.0\"?>\n<!-- c -->\n<a x=\"1\">\n  <b>text &amp; more</b>\n  <c/>\n</a>\n"
	if err := (Validator{}).Validate(src); err != nil {
		t.Errorf("Validate() = %v, want nil", err)
	}
}

func TestValidateReportsLine(t *testing.T) {
	tests := []struct {
		name string
		src  string
		line int
	}{
		{"wrong end tag", "<a>\n  <b>\n</a>", 3},
		{"unclosed element", "<a>\n  <b>", 2},
		{"two root elements", "<a/>\n<b/>", 2},
		{"text outside root", "<a/>\noops", 2},
		{"no root element", "<!-- only a comment -->", 1},
		{"undefined entity", "<a>\n&nbsp;</a>", 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validator{}.Validate(tt.src)
			var syntaxErr *filetype.SyntaxError
			if !errors.As(err, &syntaxErr) {
				t.Fatalf("Validate() = %v, want a SyntaxError", err)
			}
			if syntaxErr.Line != tt.line {
				t.Errorf("error line = %d, want %d (%v)", syntaxErr.Line, tt.line, err)
			}
		})
	}
}
