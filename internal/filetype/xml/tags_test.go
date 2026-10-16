package xml

import (
	"cmp"
	"slices"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/ixexel661/vex/internal/buffer"
	"github.com/ixexel661/vex/internal/highlight"
)

// source is a document as the editor shows it to Tags.
type source struct {
	*buffer.Buffer
	highlights *highlight.Cache
}

func (s source) State(row int) highlight.State {
	return s.highlights.State(s.Buffer, row)
}

// withCursor reads a document with the cursor marked as "|".
func withCursor(text string) (source, buffer.Position) {
	before, after, _ := strings.Cut(text, "|")
	lines := strings.Split(before, "\n")
	p := buffer.Position{Row: len(lines) - 1, Col: utf8.RuneCountInString(lines[len(lines)-1])}
	return source{buffer.New(before + after), highlight.NewCache(Highlighter{})}, p
}

// bracket puts each range of text in brackets.
func bracket(text string, ranges ...buffer.Range) string {
	lines := strings.Split(text, "\n")
	// The last range first, so the brackets don't move the others.
	slices.SortFunc(ranges, func(a, b buffer.Range) int {
		return cmp.Or(cmp.Compare(b.Start.Row, a.Start.Row), cmp.Compare(b.Start.Col, a.Start.Col))
	})
	for _, r := range ranges {
		line := []rune(lines[r.Start.Row])
		line = slices.Insert(line, r.End.Col, ']')
		line = slices.Insert(line, r.Start.Col, '[')
		lines[r.Start.Row] = string(line)
	}
	return strings.Join(lines, "\n")
}

func TestPair(t *testing.T) {
	tests := []struct {
		name, text, want string
	}{
		{"start tag", `<bo|ok id="1">x</book>`, `<[book] id="1">x</[book]>`},
		{"end tag", `<book id="1">x</bo|ok>`, `<[book] id="1">x</[book]>`},
		{"start of the name", "<|book>x</book>", "<[book]>x</[book]>"},
		{"end of the name", "<book|>x</book>", "<[book]>x</[book]>"},
		{"start of the end name", "<book>x</|book>", "<[book]>x</[book]>"},
		{"before the tag", "|<book>x</book>", ""},
		{"in the attributes", `<book |id="1">x</book>`, ""},
		{"between the tags", "<book>|</book>", ""},
		{"over several lines", "<book>\n  x\n</bo|ok>", "<[book]>\n  x\n</[book]>"},
		{"start tag over several lines", "<bo|ok\n  id=\"1\">x</book>", "<[book]\n  id=\"1\">x</[book]>"},
		{"back to a start tag over several lines", "<book\n  id=\"1\">x</bo|ok>", "<[book]\n  id=\"1\">x</[book]>"},
		{"inner of the same name", "<a><|a></a></a>", "<a><[a]></[a]></a>"},
		{"outer of the same name", "<|a><a></a></a>", "<[a]><a></a></[a]>"},
		{"back to the outer of the same name", "<a><a></a></|a>", "<[a]><a></a></[a]>"},
		{"comment", "<|a><!-- </a> --></a>", "<[a]><!-- </a> --></[a]>"},
		{"comment over several lines", "<a>\n<!--\n<a>\n-->\n</|a>", "<[a]>\n<!--\n<a>\n-->\n</[a]>"},
		{"CDATA", "<|a><![CDATA[</a>]]></a>", "<[a]><![CDATA[</a>]]></[a]>"},
		{"attribute value", `<|a x="</a>"></a>`, `<[a] x="</a>"></[a]>`},
		{"self-closing", "<|a/></a>", ""},
		{"self-closing child", "<|a><a/></a>", "<[a]><a/></[a]>"},
		{"self-closing in an element of the same name", "<a><|a/></a>", ""},
		{"unfinished", "<|a\n</a>", ""},
		{"unfinished end tag", "<|a>x</a", ""},
		{"unfinished child", "<|a><b </a>", "<[a]><b </[a]>"},
		{"no end tag, the search stops at the end of the parent", "<a><p><|a></p></a>", ""},
		{"child without an end tag", "<|p><b>x</p>", "<[p]><b>x</[p]>"},
		{"stray end tag", "<|a></b></a>", ""},
		{"empty names", "<|>x</>", "<[]>x</[]>"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src, p := withCursor(tt.text)
			name, other, ok := Tags{}.Pair(src, p)
			got := ""
			if ok {
				got = bracket(src.Text(), name, other)
			}
			if got != tt.want {
				t.Errorf("Pair() = %q, want %q", got, tt.want)
			}
			if ok && (p.Before(name.Start) || name.End.Before(p)) {
				t.Errorf("name %v does not contain the cursor %v", name, p)
			}
		})
	}
}

func TestClose(t *testing.T) {
	tests := []struct {
		name, text, want string
	}{
		{"start tag", "<a>|", "</a>"},
		{"with attributes", `<a x="1" y='>'>|`, "</a>"},
		{"over several lines", "<a\n  x=\"1\">|", "</a>"},
		{"after the XML declaration", "<?xml version=\"1.0\"?>\n<a>|", "</a>"},
		{"in an element", "<p><b>|</p>", "</b>"},
		{"in an element of the same name", "<b><p><b>|</p></b>", "</b>"},
		{"self-closing", "<a/>|", ""},
		{"end tag", "<a></a>|", ""},
		{"in an attribute value", `<a x=">|"`, ""},
		{"in a comment", "<!-- <a>| -->", ""},
		{"in CDATA", "<![CDATA[<a>|]]>", ""},
		{"in a processing instruction", `<?xml version="1.0"?>|`, ""},
		{"in a declaration", "<!DOCTYPE a>|", ""},
		{"in text", "<a>x>|", ""},
		{"in text on the next line", "<a>\nxy>|", ""},
		{"without a name", "<>|", ""},
		{"already closed", "<a>|</a>", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src, p := withCursor(tt.text)
			got, ok := Tags{}.Close(src, p)
			if got != tt.want || ok != (tt.want != "") {
				t.Errorf("Close() = %q, %v, want %q", got, ok, tt.want)
			}
		})
	}
}

func TestIsName(t *testing.T) {
	for text, want := range map[string]bool{
		"":          true,
		"book":      true,
		"x:a-b_1.2": true,
		"äöü":       true,
		" ":         false,
		">":         false,
		"/":         false,
		"=":         false,
		`"`:         false,
		"\t":        false,
		"\n":        false,
		"a b":       false,
	} {
		if got := (Tags{}).IsName(text); got != want {
			t.Errorf("IsName(%q) = %v, want %v", text, got, want)
		}
	}
}
