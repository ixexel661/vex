package xml

import (
	"errors"
	"testing"

	"github.com/ixexel661/vex/internal/config"
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
			got, err := Formatter{}.Format(tt.src, config.Format{Indent: "  "})
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
	_, err := Formatter{}.Format("<a>\n<b>\n</a>", config.Format{Indent: "  "})
	var syntaxErr *filetype.SyntaxError
	if !errors.As(err, &syntaxErr) {
		t.Fatalf("Format() error = %v, want a SyntaxError", err)
	}
	if syntaxErr.Line != 3 {
		t.Errorf("error line = %d, want 3", syntaxErr.Line)
	}
}

func TestFormatOptions(t *testing.T) {
	const indent = config.Indent("  ")
	tests := []struct {
		name string
		opts config.Format
		src  string
		want string
	}{
		{
			name: "empty elements close themselves",
			opts: config.Format{Indent: indent, SelfCloseEmpty: true},
			src:  `<a><b></b><c x="1" ></c><d>  </d><e/><f>text</f></a>`,
			want: "<a>\n  <b/>\n  <c x=\"1\"/>\n  <d/>\n  <e/>\n  <f>text</f>\n</a>\n",
		},
		{
			name: "blank lines between elements stay once",
			opts: config.Format{Indent: indent, KeepBlankLines: true},
			src:  "<a>\n\n  <b/>\n\n\n  <c/>\n  <d/>\n\n</a>\n",
			want: "<a>\n  <b/>\n\n  <c/>\n  <d/>\n</a>\n",
		},
		{
			name: "blank lines go without the option",
			opts: config.Format{Indent: indent},
			src:  "<a>\n\n  <b/>\n\n\n  <c/>\n  <d/>\n\n</a>\n",
			want: "<a>\n  <b/>\n  <c/>\n  <d/>\n</a>\n",
		},
		{
			name: "blank line after the declaration",
			opts: config.Format{Indent: indent, KeepBlankLines: true},
			src:  "<?xml version=\"1.0\"?>\n\n<a/>\n",
			want: "<?xml version=\"1.0\"?>\n\n<a/>\n",
		},
		{
			name: "long tags put each attribute on a line",
			opts: config.Format{Indent: indent, WrapTagsLongerThan: 20},
			src:  `<a><b x="1" y="2"/><c name="a long value" id="c1"><d/></c></a>`,
			want: "<a>\n  <b x=\"1\" y=\"2\"/>\n  <c\n    name=\"a long value\"\n    id=\"c1\">\n    <d/>\n  </c>\n</a>\n",
		},
		{
			name: "a tag that fits goes on one line",
			opts: config.Format{Indent: indent, WrapTagsLongerThan: 20},
			src:  "<a\n   x = \"1\"\n   y='2'>\n<b/></a>",
			want: "<a x=\"1\" y='2'>\n  <b/>\n</a>\n",
		},
		{
			name: "text stays and values may contain >",
			opts: config.Format{Indent: indent, WrapTagsLongerThan: 10},
			src:  `<a><t lang="en" note="a > b">Go</t></a>`,
			want: "<a>\n  <t\n    lang=\"en\"\n    note=\"a > b\">Go</t>\n</a>\n",
		},
		{
			name: "values may contain the other quote",
			opts: config.Format{Indent: indent, WrapTagsLongerThan: 10},
			src:  `<a><q x="it's" y='say "hi"'/></a>`,
			want: "<a>\n  <q\n    x=\"it's\"\n    y='say \"hi\"'/>\n</a>\n",
		},
		{
			name: "self-closing and wrapped",
			opts: config.Format{Indent: indent, SelfCloseEmpty: true, WrapTagsLongerThan: 10},
			src:  `<a><e first="1" second="2"></e></a>`,
			want: "<a>\n  <e\n    first=\"1\"\n    second=\"2\"/>\n</a>\n",
		},
		{
			name: "tags without attributes stay",
			opts: config.Format{Indent: indent, WrapTagsLongerThan: 5},
			src:  "<a><averyverylongname/></a>",
			want: "<a>\n  <averyverylongname/>\n</a>\n",
		},
		{
			name: "tabs",
			opts: config.Format{Indent: "\t", WrapTagsLongerThan: 5},
			src:  `<a><b x="1" y="2"/></a>`,
			want: "<a>\n\t<b\n\t\tx=\"1\"\n\t\ty=\"2\"/>\n</a>\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Formatter{}.Format(tt.src, tt.opts)
			if err != nil {
				t.Fatalf("Format() error: %v", err)
			}
			if got != tt.want {
				t.Errorf("Format() =\n%s\nwant\n%s", got, tt.want)
			}
		})
	}
}
