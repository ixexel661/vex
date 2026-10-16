package xml

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/ixexel661/vex/internal/buffer"
	"github.com/ixexel661/vex/internal/filetype"
)

type jump func(filetype.Source, buffer.Position) (buffer.Position, bool)

// jumped returns text with the cursor "|" where jump puts it, or "" if it
// stays.
func jumped(text string, jump jump) string {
	src, p := withCursor(text)
	to, ok := jump(src, p)
	if !ok {
		return ""
	}
	lines := strings.Split(strings.Replace(text, "|", "", 1), "\n")
	if to.Row < 0 || to.Row >= len(lines) || to.Col < 0 || to.Col > len([]rune(lines[to.Row])) {
		return fmt.Sprintf("outside of the text: %v", to)
	}
	lines[to.Row] = string(slices.Insert([]rune(lines[to.Row]), to.Col, '|'))
	return strings.Join(lines, "\n")
}

type jumpTest struct {
	name, text, want string
}

func testJump(t *testing.T, jump jump, tests []jumpTest) {
	t.Helper()
	for _, tt := range tests {
		if got := jumped(tt.text, jump); got != tt.want {
			t.Errorf("%s: %q goes to %q, want %q", tt.name, tt.text, got, tt.want)
		}
	}
}

func TestNextTag(t *testing.T) {
	testJump(t, Tags{}.NextTag, []jumpTest{
		{"from text", "a|b<x>c</x>", "ab|<x>c</x>"},
		{"at a tag", "|<a>x</a>", "<a>x|</a>"},
		{"from inside a tag", "<ab|c>x</abc>", "<abc>x|</abc>"},
		{"to an empty tag", "<a|><b/></a>", "<a>|<b/></a>"},
		{"over lines", "<a>|\n  <b/>", "<a>\n  |<b/>"},
		{"skips comments and CDATA", "<a>|<!-- <x> --><![CDATA[<y>]]><b/></a>", "<a><!-- <x> --><![CDATA[<y>]]>|<b/></a>"},
		{"skips attribute values", `<b| x="<c>"/><d/>`, `<b x="<c>"/>|<d/>`},
		{"in the last tag", "<a></a|>", ""},
		{"after the last tag", "<a></a>|", ""},
	})
}

func TestPrevTag(t *testing.T) {
	testJump(t, Tags{}.PrevTag, []jumpTest{
		{"from text", "<a>x|</a>", "|<a>x</a>"},
		{"at a tag", "<a><b/>|</a>", "<a>|<b/></a>"},
		{"from inside an end tag", "<a></a|>", "<a>|</a>"},
		{"over lines", "<a>\n  |x", "|<a>\n  x"},
		{"skips comments", "<a><!-- <x> -->|</a>", "|<a><!-- <x> --></a>"},
		{"at the first tag", "|<a/>", ""},
		{"before the first tag", "x|<a/>", ""},
	})
}

func TestParent(t *testing.T) {
	testJump(t, Tags{}.Parent, []jumpTest{
		{"from text", "<a><b>x|y</b></a>", "<a>|<b>xy</b></a>"},
		{"at the start tag", "<a>|<b>x</b></a>", "|<a><b>x</b></a>"},
		{"at the end tag", "<a><b>x|</b></a>", "<a>|<b>x</b></a>"},
		{"in the end tag", "<a><b>x</b|></a>", "<a>|<b>x</b></a>"},
		{"after the end tag", "<a><b>x</b>|</a>", "|<a><b>x</b></a>"},
		{"after siblings", "<a><b><c/></b><d>x</d>|</a>", "|<a><b><c/></b><d>x</d></a>"},
		{"in the attributes", `<a><b |x="1">y</b></a>`, `<a>|<b x="1">y</b></a>`},
		{"in a start tag over lines", "<a>\n  <b\n    x=\"1\"|\n    y=\"2\">\n  </b>\n</a>", "<a>\n  |<b\n    x=\"1\"\n    y=\"2\">\n  </b>\n</a>"},
		{"in an empty tag", `<a><b x|="1"/></a>`, `<a>|<b x="1"/></a>`},
		{"after an empty tag", "<a><b/>|</a>", "|<a><b/></a>"},
		{"skips comments", "<a><!-- <b> -->|</a>", "|<a><!-- <b> --></a>"},
		{"at the root", "|<a></a>", ""},
		{"before the root", "<?xml version=\"1.0\"?>|\n<a/>", ""},
		{"after the root", "<a></a>|", ""},
	})
}

func TestParentUpToTheRoot(t *testing.T) {
	text := "<a>\n  <b>\n    <c>x|</c>\n  </b>\n</a>"
	for _, want := range []string{
		"<a>\n  <b>\n    |<c>x</c>\n  </b>\n</a>",
		"<a>\n  |<b>\n    <c>x</c>\n  </b>\n</a>",
		"|<a>\n  <b>\n    <c>x</c>\n  </b>\n</a>",
		"",
	} {
		got := jumped(text, Tags{}.Parent)
		if got != want {
			t.Fatalf("%q goes to %q, want %q", text, got, want)
		}
		text = got
	}
}

func TestFirstChild(t *testing.T) {
	testJump(t, Tags{}.FirstChild, []jumpTest{
		{"at a start tag", "|<a><b/></a>", "<a>|<b/></a>"},
		{"from text in the element", "<a>\n  |\n  <b>x</b>\n</a>", "<a>\n  \n  |<b>x</b>\n</a>"},
		{"skips text and comments", "|<a>text<!-- <x> --><b/></a>", "<a>text<!-- <x> -->|<b/></a>"},
		{"in the start tag", `<a |x="1"><b/></a>`, `<a x="1">|<b/></a>`},
		{"at the end tag", "<a><b/>|</a>", "<a>|<b/></a>"},
		{"without children", "|<a>x</a>", ""},
		{"at an empty tag", "<a>|<b/><c/></a>", ""},
		{"from the declaration to the root", "<?xml version=\"1.0\"?>|\n<a><b/></a>", "<?xml version=\"1.0\"?>\n|<a><b/></a>"},
		{"without elements", "x|", ""},
	})
}

func TestParentAndFirstChildGoBackAndForth(t *testing.T) {
	text := "<a>\n  <b>x</b>\n  <c/>\n</a>"
	src, _ := withCursor(text)
	child := buffer.Position{Row: 1, Col: 2}
	parent, ok := Tags{}.Parent(src, child)
	if !ok || parent != (buffer.Position{}) {
		t.Fatalf("Parent(%v) = %v, %v, want {0 0}", child, parent, ok)
	}
	if got, ok := (Tags{}).FirstChild(src, parent); !ok || got != child {
		t.Errorf("FirstChild(%v) = %v, %v, want %v", parent, got, ok, child)
	}
}
