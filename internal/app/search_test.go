package app

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/ixexel661/vex/internal/buffer"
	"github.com/ixexel661/vex/internal/highlight"
	"github.com/ixexel661/vex/internal/mode"
	"github.com/ixexel661/vex/internal/theme"
)

const books = "<books>\n  <book>\n    <title>Go</title>\n  </book>\n  <book>\n    <title>XML</title>\n  </book>\n</books>\n"

var (
	ctrlF     = tea.KeyPressMsg{Code: 'f', Mod: tea.ModCtrl}
	tab       = tea.KeyPressMsg{Code: tea.KeyTab}
	enter     = tea.KeyPressMsg{Code: tea.KeyEnter}
	backspace = tea.KeyPressMsg{Code: tea.KeyBackspace}
	down      = tea.KeyPressMsg{Code: tea.KeyDown}
	up        = tea.KeyPressMsg{Code: tea.KeyUp}
)

// typing returns the key presses that type text.
func typing(text string) []tea.KeyPressMsg {
	var keys []tea.KeyPressMsg
	for _, r := range text {
		keys = append(keys, key(string(r)))
	}
	return keys
}

func at(row, col int) buffer.Position {
	return buffer.Position{Row: row, Col: col}
}

func searchFrom(t *testing.T, row int, query string) *Model {
	t.Helper()
	m, _ := newModel(t, books)
	for range row {
		press(m, down)
	}
	press(m, ctrlF)
	press(m, typing(query)...)
	return m
}

func TestSearchJumpsToFirstMatchFromCursor(t *testing.T) {
	m := searchFrom(t, 3, "title")
	if got := m.editor.Cursor(); got != at(5, 5) {
		t.Errorf("cursor = %v, want the first match after it at {5 5}", got)
	}
	if got, want := m.messageLine(), " TEXT   title  3/4  "+searchHints; got != want {
		t.Errorf("search bar = %q, want %q", got, want)
	}
	if c := m.View().Cursor; c.X != len(" TEXT   title") || c.Y != 23 {
		t.Errorf("terminal cursor at %d,%d, want it after the query at %d,23", c.X, c.Y, len(" TEXT   title"))
	}
}

func TestSearchArrowsSelectMatches(t *testing.T) {
	m := searchFrom(t, 3, "title")
	for _, step := range []struct {
		key  tea.KeyPressMsg
		want buffer.Position
	}{{down, at(5, 16)}, {down, at(2, 5)}, {up, at(5, 16)}} {
		press(m, step.key)
		if got := m.editor.Cursor(); got != step.want {
			t.Fatalf("cursor after %s = %v, want %v", step.key, got, step.want)
		}
	}
}

func TestSearchEnterStaysAtMatch(t *testing.T) {
	m := searchFrom(t, 3, "title")
	press(m, down, enter)
	if m.searching {
		t.Error("enter did not close the search")
	}
	if got := m.editor.Cursor(); got != at(5, 16) {
		t.Errorf("cursor = %v, want the match at {5 16}", got)
	}
}

func TestSearchEscGoesBack(t *testing.T) {
	m := searchFrom(t, 3, "title")
	press(m, down, esc)
	if m.searching {
		t.Error("esc did not close the search")
	}
	if got := m.editor.Cursor(); got != at(3, 0) {
		t.Errorf("cursor = %v, want the start of the search at {3 0}", got)
	}
}

func TestSearchWithoutMatches(t *testing.T) {
	m := searchFrom(t, 3, "nothing")
	if got := m.editor.Cursor(); got != at(3, 0) {
		t.Errorf("cursor = %v, want it to stay at {3 0}", got)
	}
	if got := m.messageLine(); !strings.Contains(got, "no matches") {
		t.Errorf("search bar = %q, want it to say there are no matches", got)
	}
}

func TestSearchTypesLetters(t *testing.T) {
	m, _ := newModel(t, books)
	press(m, ctrlF)
	if isQuit(press(m, key("q"))) {
		t.Fatal("q quit while searching")
	}
	if got := m.search.Query(); got != "q" {
		t.Errorf("query = %q, want %q", got, "q")
	}
	if got := m.editor.Document().Buffer.Text(); got != books {
		t.Errorf("text changed to %q", got)
	}
}

func TestSearchKeepsMode(t *testing.T) {
	m, _ := newModel(t, books)
	press(m, key("i"), ctrlF)
	press(m, typing("title")...)
	press(m, enter)
	if m.mode != mode.Insert {
		t.Errorf("mode = %v, want INSERT", m.mode)
	}
	if got := m.editor.Document().Buffer.Text(); got != books {
		t.Errorf("text changed to %q", got)
	}
}

func TestSearchTabSwitchesToFuzzyList(t *testing.T) {
	m := searchFrom(t, 3, "title")
	press(m, tab)
	if got := m.editor.Cursor(); got != at(2, 5) {
		t.Errorf("cursor = %v, want the best match at {2 5}", got)
	}
	if got := strings.Count(m.editor.View(), "\n") + 1; got >= 22 {
		t.Errorf("editor has %d rows, want fewer than 22 to make room for the list", got)
	}
	view := m.View().Content
	if got := strings.Count(view, "\n") + 1; got != 24 {
		t.Errorf("view has %d rows, want 24", got)
	}
	for _, want := range []string{" FUZZY   title  1/2", "3  <title>Go</title>", "6  <title>XML</title>"} {
		if !strings.Contains(view, want) {
			t.Errorf("view does not show %q:\n%s", want, view)
		}
	}

	press(m, tab)
	if got := strings.Count(m.editor.View(), "\n") + 1; got != 22 {
		t.Errorf("editor has %d rows after switching back, want 22", got)
	}
	if got := m.editor.Cursor(); got != at(5, 5) {
		t.Errorf("cursor = %v, want the first match from the start again at {5 5}", got)
	}
}

func TestSearchMarksMatches(t *testing.T) {
	brackets := func(open, close string) lipgloss.Style {
		return lipgloss.NewStyle().Transform(func(s string) string { return open + s + close })
	}
	th := theme.Theme{Syntax: map[highlight.Kind]lipgloss.Style{
		highlight.Match:        brackets("[", "]"),
		highlight.CurrentMatch: brackets("{", "}"),
	}}
	m, _ := newModelWithTheme(t, books, th)
	expect := func(when string, rows map[int]string) {
		t.Helper()
		lines := strings.Split(m.editor.View(), "\n")
		for n, want := range rows {
			if !strings.HasSuffix(lines[n], want) {
				t.Errorf("%s: row %d = %q, want %q", when, n, lines[n], want)
			}
		}
	}

	press(m, ctrlF)
	press(m, typing("title")...)
	expect("text search", map[int]string{2: "<{title}>Go</[title]>", 5: "<[title]>XML</[title]>"})
	press(m, tab)
	expect("fuzzy search", map[int]string{2: "<{title}>Go</title>", 5: "<title>XML</title>"})
	press(m, enter)
	expect("after the search", map[int]string{2: "<title>Go</title>", 5: "<title>XML</title>"})
}

func TestPasteGoesToSearch(t *testing.T) {
	m, _ := newModel(t, books)
	press(m, ctrlF)
	m.Update(tea.PasteMsg{Content: "title\r\nmore"})
	if got := m.search.Query(); got != "title" {
		t.Errorf("query = %q, want the first line of the pasted text", got)
	}
	if got := m.editor.Document().Buffer.Text(); got != books {
		t.Errorf("text changed to %q", got)
	}
}

func TestSearchStartsWithLastQuery(t *testing.T) {
	m := searchFrom(t, 0, "title")
	press(m, enter, ctrlF)
	if got := m.search.Query(); got != "title" {
		t.Fatalf("query = %q, want the last one", got)
	}
	press(m, down, enter)
	if got := m.editor.Cursor(); got != at(2, 15) {
		t.Errorf("cursor = %v, want the match after the last one at {2 15}", got)
	}

	press(m, ctrlF)
	press(m, typing("book")...)
	if got := m.search.Query(); got != "book" {
		t.Errorf("query = %q, want typing to replace the last one", got)
	}

	press(m, esc, ctrlF, backspace)
	if got := m.search.Query(); got != "" {
		t.Errorf("query = %q, want backspace to clear the last one", got)
	}
}
