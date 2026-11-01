package app

import (
	"fmt"

	tea "charm.land/bubbletea/v2"

	"github.com/ixexel661/vex/internal/printable"
)

const (
	schemaHints = "↑↓ select · tab complete · ctrl+n add · ctrl+d remove · enter save · esc cancel"
	// labelWidth is the width of the marker and the label of a field, as in
	// "▸ Namespace  ".
	labelWidth = 13
	// minFieldWidth is the width of a field in a box that has room for it.
	minFieldWidth = 40
)

func (m *Model) schemaBox() string {
	return m.popup("Schemas", m.schemaLines())
}

// schemaLines are the lines in the schema popup, with a blank line between
// two schemas.
func (m *Model) schemaLines() []string {
	p := m.schemaPopup
	width := m.fieldWidth()
	var lines []string
	for i, f := range p.fields {
		if i > 0 && i%2 == 0 {
			lines = append(lines, "")
		}
		label, marker := "Namespace", " "
		if i%2 == 1 {
			label = "File"
		}
		if i == p.focused {
			marker = "▸"
		}
		text, _ := f.view(width)
		lines = append(lines, fmt.Sprintf("%s %-9s  ", marker, label)+text)
	}
	lines = append(lines, "")
	if p.err != nil {
		lines = append(lines, m.theme.Error.Render(printable.String(p.err.Error())))
	}
	return append(lines, m.theme.Hint.Render(m.schemasLocation()))
}

// fieldWidth is as wide as the longest text needs, but fits on the screen.
func (m *Model) fieldWidth() int {
	width := minFieldWidth
	for _, f := range m.schemaPopup.fields {
		width = max(width, len(f.text)+1)
	}
	// The popup border takes two columns on each side.
	return max(min(width, m.width-4-labelWidth), 1)
}

// schemasLocation says where the schemas are stored.
func (m *Model) schemasLocation() string {
	if m.editor.Document().Path == "" || m.schemasPath == "" {
		return "for this session"
	}
	return homeRelative(m.schemasPath)
}

// schemaCursor puts a bar cursor in the focused field.
func (m *Model) schemaCursor() *tea.Cursor {
	p := m.schemaPopup
	x, y := m.boxPosition(m.schemaBox())
	_, col := p.fields[p.focused].view(m.fieldWidth())
	// The border line and a blank line between two schemas come before it.
	c := tea.NewCursor(x+2+labelWidth+col, y+1+p.focused+p.focused/2)
	c.Shape = tea.CursorBar
	return c
}
