package app

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"charm.land/lipgloss/v2"

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
	query := printable(m.search.Query())
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

func (m *Model) listHeight() int {
	if !m.searching || !m.search.Matcher().Ranked() {
		return 0
	}
	rows := min(maxListRows, m.height/3-1)
	if rows < 1 {
		return 0
	}
	return rows + 1
}

func (m *Model) matchList(height int) string {
	matches, selected := m.search.Matches(), m.search.Selected()
	visible := height - 1
	first := min(max(selected-visible/2, 0), max(len(matches)-visible, 0))
	numberWidth := len(strconv.Itoa(m.editor.Document().Buffer.LineCount()))

	rows := make([]string, height)
	rows[0] = m.theme.Filler.Render(strings.Repeat("─", m.width))
	for y := 1; y < height; y++ {
		if i := first + y - 1; i < len(matches) {
			rows[y] = m.listRow(matches[i], numberWidth, i == selected)
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
		row += text.Render(printable(string(line[pos:start]))) + matched.Render(printable(string(line[start:end])))
		pos = end
	}
	row += text.Render(printable(string(line[pos:])))
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

func printable(s string) string {
	return strings.Map(func(r rune) rune {
		switch {
		case r == '\t':
			return ' '
		case unicode.IsControl(r):
			return '?'
		default:
			return r
		}
	}, s)
}
