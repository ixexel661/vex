package app

import (
	"fmt"
	"strconv"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/ixexel661/vex/internal/printable"
	"github.com/ixexel661/vex/internal/search"
)

const (
	searchHints = "tab switch · ↑↓ select · enter jump · esc cancel"
	maxListRows = 8
)

// searchBar takes the place of the message line while searching.
func (m *Model) searchBar() string {
	parts := []string{m.searchInput()}
	if count := m.matchCount(); count != "" {
		parts = append(parts, count)
	}
	return strings.Join(append(parts, m.theme.Hint.Render(searchHints)), "  ")
}

func (m *Model) searchInput() string {
	query := printable.String(m.search.Query())
	if m.replaceQuery {
		query = m.theme.Selected.Render(query)
	}
	return badge(m.theme.SearchMode, strings.ToUpper(m.search.Matcher().Name())) + "  " + query
}

func (m *Model) matchCount() string {
	switch {
	case m.search.Query() == "":
		return ""
	case m.search.Selected() < 0:
		return m.theme.Error.Render("no matches")
	default:
		return m.theme.Message.Render(fmt.Sprintf("%d/%d", m.search.Selected()+1, len(m.search.Matches())))
	}
}

// listHeight is the height of the list above the status bar: the matches
// of a search that ranks them, or the problems in the text.
func (m *Model) listHeight() int {
	count := 0
	switch {
	case m.problemList != nil:
		count = len(m.editor.Problems())
	case m.searching && m.search.Matcher().Ranked():
		count = maxListRows
	}
	rows := min(maxListRows, m.height/3-1, count)
	if rows < 1 {
		return 0
	}
	return rows + 1
}

func (m *Model) listView(height int) string {
	if m.problemList != nil {
		return m.problemListView(height)
	}
	return m.matchList(height)
}

func (m *Model) matchList(height int) string {
	matches := m.search.Matches()
	numberWidth := len(strconv.Itoa(m.editor.Document().Buffer.LineCount()))
	return m.list(height, len(matches), m.search.Selected(), func(i int, selected bool) string {
		return m.listRow(matches[i], numberWidth, selected)
	})
}

// list draws count rows under a line, as many as fit in height, scrolled so
// the selected one shows.
func (m *Model) list(height, count, selected int, row func(i int, selected bool) string) string {
	visible := height - 1
	first := min(max(selected-visible/2, 0), max(count-visible, 0))
	rows := make([]string, height)
	rows[0] = m.theme.Filler.Render(strings.Repeat("─", m.width))
	for y := 1; y < height; y++ {
		if i := first + y - 1; i < count {
			rows[y] = row(i, i == selected)
		}
	}
	return strings.Join(rows, "\n")
}

func (m *Model) listRow(match search.Match, numberWidth int, selected bool) string {
	text, number, matched := lipgloss.NewStyle(), m.theme.LineNumber, m.theme.MatchedRune
	if selected {
		text, number, matched = m.theme.Selected, m.theme.Selected, matched.Inherit(m.theme.Selected)
	}
	line := []rune(m.editor.Document().Buffer.Line(match.Row))
	pos := indentation(line)
	row := number.Render(fmt.Sprintf("%*d  ", numberWidth, match.Row+1))
	for _, r := range match.Ranges {
		start, end := max(r.Start, pos), max(r.End, pos)
		row += text.Render(printable.String(string(line[pos:start]))) + matched.Render(printable.String(string(line[start:end])))
		pos = end
	}
	row += text.Render(printable.String(string(line[pos:])))
	if selected {
		row += text.Render(strings.Repeat(" ", max(m.width-lipgloss.Width(row), 0)))
	}
	return row
}

// indentation returns the number of spaces and tabs a line starts with.
func indentation(line []rune) int {
	n := 0
	for n < len(line) && (line[n] == ' ' || line[n] == '\t') {
		n++
	}
	return n
}
