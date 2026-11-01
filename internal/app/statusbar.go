package app

import (
	"errors"
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/ixexel661/vex/internal/filetype"
	"github.com/ixexel661/vex/internal/mode"
)

func (m *Model) statusBar() string {
	doc := m.editor.Document()
	cursor := m.editor.Cursor()

	modeBadge := badge(m.theme.NormalMode, m.mode.String())
	if m.mode == mode.Insert {
		modeBadge = badge(m.theme.InsertMode, m.mode.String())
	}
	name := " " + doc.Name()
	if doc.Modified() {
		name += " [+]"
	}
	validity := m.validity()
	position := fmt.Sprintf("  %s  Ln %d, Col %d ", doc.Type.Name, cursor.Row+1, cursor.Col+1)

	used := lipgloss.Width(modeBadge) + lipgloss.Width(name) + lipgloss.Width(validity) + lipgloss.Width(position)
	gap := strings.Repeat(" ", max(m.width-used, 1))
	return modeBadge + m.theme.StatusBar.Render(name+gap) + validity + m.theme.StatusBar.Render(position)
}

// badge shows text as a label with a space on each side.
func badge(style lipgloss.Style, text string) string {
	return style.Render(" " + text + " ")
}

func (m *Model) validity() string {
	doc := m.editor.Document()
	if doc.Type.Validator == nil && doc.Type.Schemas == nil {
		return ""
	}
	v := m.validation
	switch {
	case v.running:
		return m.theme.Hint.Render("… checking")
	case v.err != nil:
		var syntaxErr *filetype.SyntaxError
		if errors.As(v.err, &syntaxErr) {
			return m.theme.Invalid.Render(fmt.Sprintf("✗ line %d", syntaxErr.Line))
		}
		return m.theme.Invalid.Render("✗ invalid")
	case v.schemaErr != nil:
		return m.theme.Invalid.Render("✗ schema")
	case v.schema:
		return m.schemaValidity(doc.Type.Schemas.Name())
	default:
		return m.theme.Valid.Render("✓ valid")
	}
}

// schemaValidity counts the problems against the schemas.
func (m *Model) schemaValidity(kind string) string {
	n := len(m.editor.Problems())
	switch {
	case n == 0:
		return m.theme.Valid.Render("✓ " + kind + " valid")
	case n == 1:
		return m.theme.Invalid.Render("✗ 1 " + kind + " error")
	case n >= filetype.MaxProblems:
		return m.theme.Invalid.Render(fmt.Sprintf("✗ %d+ %s errors", n, kind))
	default:
		return m.theme.Invalid.Render(fmt.Sprintf("✗ %d %s errors", n, kind))
	}
}
