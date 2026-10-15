package xml

import (
	"strings"
	"testing"

	"github.com/ixexel661/vex/internal/highlight"
)

// colorize renders the spans of each line as "text:Kind" pairs, which makes
// test expectations easy to read.
func colorize(lines ...string) []string {
	names := map[highlight.Kind]string{
		highlight.Plain:       "plain",
		highlight.Punctuation: "punct",
		highlight.Tag:         "tag",
		highlight.Attribute:   "attr",
		highlight.String:      "string",
		highlight.Comment:     "comment",
		highlight.Meta:        "meta",
		highlight.Entity:      "entity",
		highlight.CData:       "cdata",
	}
	var out []string
	state := highlight.State(0)
	for _, line := range lines {
		runes := []rune(line)
		var spans []highlight.Span
		spans, state = Highlighter{}.Highlight(runes, state)
		var parts []string
		for _, s := range spans {
			parts = append(parts, string(runes[s.Start:s.End])+":"+names[s.Kind])
		}
		out = append(out, strings.Join(parts, " "))
	}
	return out
}

func TestHighlight(t *testing.T) {
	tests := []struct {
		name  string
		lines []string
		want  []string
	}{
		{
			name:  "element with attribute",
			lines: []string{`<book id="1">Go</book>`},
			want:  []string{`<:punct book:tag  :plain id:attr =:punct "1":string >:punct Go:plain </:punct book:tag >:punct`},
		},
		{
			name:  "self-closing element and entity",
			lines: []string{`<a/>x &amp; y`},
			want:  []string{`<:punct a:tag />:punct x :plain &amp;:entity  y:plain`},
		},
		{
			name:  "declaration and comment",
			lines: []string{`<?xml version="1.0"?><!-- hi -->`},
			want:  []string{`<?xml version="1.0"?>:meta <!-- hi -->:comment`},
		},
		{
			name:  "comment over several lines",
			lines: []string{"<!-- start", "middle", "end --><a>"},
			want:  []string{"<!-- start:comment", "middle:comment", "end -->:comment <:punct a:tag >:punct"},
		},
		{
			name:  "attributes over several lines",
			lines: []string{`<a x="1`, `2" y='3'>`},
			want:  []string{`<:punct a:tag  :plain x:attr =:punct "1:string`, `2":string  :plain y:attr =:punct '3':string >:punct`},
		},
		{
			name:  "CDATA",
			lines: []string{"<![CDATA[a < b]]>"},
			want:  []string{"<![CDATA[a < b]]>:cdata"},
		},
		{
			name:  "unclosed tag recovers",
			lines: []string{"<a", "<b>"},
			want:  []string{"<:punct a:tag", "<:punct b:tag >:punct"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := colorize(tt.lines...)
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Errorf("line %d:\ngot  %s\nwant %s", i, got[i], tt.want[i])
				}
			}
		})
	}
}
