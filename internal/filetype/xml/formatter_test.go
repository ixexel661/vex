package xml

import (
	"errors"
	"testing"

	"github.com/ixexel661/vex/internal/filetype"
)

func TestFormat(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want string
	}{
		{
			name: "nested elements",
			src:  `<?xml version="1.0"?><catalog><book id="1"><title>Go</title><price>10</price></book></catalog>`,
			want: "<?xml version=\"1.0\"?>\n<catalog>\n  <book id=\"1\">\n    <title>Go</title>\n    <price>10</price>\n  </book>\n</catalog>\n",
		},
		{
			name: "empty and self-closing elements",
			src:  "<a><b/><c></c><d>   </d></a>",
			want: "<a>\n  <b/>\n  <c></c>\n  <d>   </d>\n</a>\n",
		},
		{
			name: "comments and CDATA",
			src:  "<a><!-- note --><b><![CDATA[x < y]]></b></a>",
			want: "<a>\n  <!-- note -->\n  <b><![CDATA[x < y]]></b>\n</a>\n",
		},
		{
			name: "mixed content is kept",
			src:  "<a>\n<p>Hello <b>World</b>!</p>\n</a>",
			want: "<a>\n  <p>Hello <b>World</b>!</p>\n</a>\n",
		},
		{
			name: "text and entities are kept",
			src:  "<a><b>  x &amp; &#169;  </b></a>",
			want: "<a>\n  <b>  x &amp; &#169;  </b>\n</a>\n",
		},
		{
			name: "re-indents existing indentation",
			src:  "<a>\n        <b>\n                <c/>\n        </b>\n</a>\n",
			want: "<a>\n  <b>\n    <c/>\n  </b>\n</a>\n",
		},
		{
			name: "other declared encoding",
			src:  `<?xml version="1.0" encoding="ISO-8859-1"?><a/>`,
			want: "<?xml version=\"1.0\" encoding=\"ISO-8859-1\"?>\n<a/>\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Formatter{}.Format(tt.src, "  ")
			if err != nil {
				t.Fatalf("Format() error: %v", err)
			}
			if got != tt.want {
				t.Errorf("Format() =\n%s\nwant\n%s", got, tt.want)
			}
		})
	}
}

func TestFormatRejectsInvalidXML(t *testing.T) {
	_, err := Formatter{}.Format("<a>\n<b>\n</a>", "  ")
	var syntaxErr *filetype.SyntaxError
	if !errors.As(err, &syntaxErr) {
		t.Fatalf("Format() error = %v, want a SyntaxError", err)
	}
	if syntaxErr.Line != 3 {
		t.Errorf("error line = %d, want 3", syntaxErr.Line)
	}
}
