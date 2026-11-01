package app

import (
	"fmt"
	"strconv"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/ixexel661/vex/internal/filetype"
	"github.com/ixexel661/vex/internal/printable"
)

const problemHints = "↑↓ select · enter jump · esc back"

// problemBar takes the place of the message line while the problems are
// listed.
func (m *Model) problemBar() string {
	count := fmt.Sprintf("%d/%d", m.problemList.selected+1, len(m.editor.Problems()))
	return m.theme.Message.Render(count) + "  " + m.theme.Hint.Render(problemHints)
}

func (m *Model) problemListView(height int) string {
	problems := m.editor.Problems()
	numberWidth := len(strconv.Itoa(m.editor.Document().Buffer.LineCount()))
	return m.list(height, len(problems), m.problemList.selected, func(i int, selected bool) string {
		return m.problemRow(problems[i], numberWidth, selected)
	})
}

func (m *Model) problemRow(p filetype.Problem, numberWidth int, selected bool) string {
	text, number := lipgloss.NewStyle(), m.theme.LineNumber
	if selected {
		text, number = m.theme.Selected, m.theme.Selected
	}
	row := number.Render(fmt.Sprintf("%*d  ", numberWidth, p.Range.Start.Row+1)) + text.Render(printable.String(p.Message))
	if selected {
		row += text.Render(strings.Repeat(" ", max(m.width-lipgloss.Width(row), 0)))
	}
	return row
}
