package app

import (
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"

	"github.com/ixexel661/vex/internal/editor"
	"github.com/ixexel661/vex/internal/highlight"
	"github.com/ixexel661/vex/internal/search"
)

func (m *Model) OpenSearch() {
	m.searching = true
	m.replaceQuery = m.search.Query() != ""
	m.search.Start(m.editor.Cursor())
	m.showMatch()
}

func (m *Model) searchKey(msg tea.KeyPressMsg) {
	switch msg.String() {
	case "enter":
		m.closeSearch(true)
	case "esc", "ctrl+c":
		m.closeSearch(false)
	case "up":
		m.search.Move(-1)
		m.showMatch()
	case "down":
		m.search.Move(1)
		m.showMatch()
	case "tab":
		m.search.NextMatcher()
		m.showMatch()
	case "backspace":
		m.deleteInSearch()
	default:
		if msg.Text != "" {
			m.typeInSearch(msg.Text)
		}
	}
}

func (m *Model) typeInSearch(text string) {
	m.editQuery(func(query string) string {
		return query + text
	})
}

func (m *Model) deleteInSearch() {
	m.editQuery(func(query string) string {
		_, size := utf8.DecodeLastRuneInString(query)
		return query[:len(query)-size]
	})
}

func (m *Model) editQuery(change func(query string) string) {
	query := m.search.Query()
	if m.replaceQuery {
		query = ""
		m.replaceQuery = false
	}
	m.search.SetQuery(change(query))
	m.showMatch()
}

func (m *Model) closeSearch(accept bool) {
	if !accept {
		m.editor.MoveTo(m.search.Origin())
	}
	m.searching = false
	m.editor.SetMarks(nil)
}

func (m *Model) showMatch() {
	selected := m.search.Selected()
	if selected < 0 {
		m.editor.SetMarks(nil)
		m.editor.MoveTo(m.search.Origin())
		return
	}
	var marks []editor.Mark
	for i, match := range m.search.Matches() {
		switch {
		case i == selected:
			marks = append(marks, marksOf(match, highlight.CurrentMatch)...)
		case !m.search.Matcher().Ranked():
			marks = append(marks, marksOf(match, highlight.Match)...)
		}
	}
	m.editor.SetMarks(marks)
	m.editor.MoveTo(m.search.Matches()[selected].Position())
}

func marksOf(match search.Match, kind highlight.Kind) []editor.Mark {
	var marks []editor.Mark
	for _, r := range match.Ranges {
		marks = append(marks, editor.Mark{Row: match.Row, Span: highlight.Span{Start: r.Start, End: r.End, Kind: kind}})
	}
	return marks
}
