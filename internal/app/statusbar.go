package app

import (
	"errors"
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/ixexel661/vex/internal/filetype"
	"github.com/ixexel661/vex/internal/mode"
)

// statusBar shows the mode, the file, its validity, its type and the cursor
// position. Each part is styled on its own, because a style reset inside a
// styled string would end the bar's background early.
func (m *Model) statusBar() string {
	doc := m.editor.Document()
	cursor := m.editor.Cursor()

	badge := m.theme.NormalMode.Render(m.mode.String())
	if m.mode == mode.Insert {
		badge = m.theme.InsertMode.Render(m.mode.String())
	}
	name := " " + doc.Name()
	if doc.Modified() {
		name += " [+]"
	}
	validity := m.validity()
	position := fmt.Sprintf("  %s  Ln %d, Col %d ", doc.Type.Name, cursor.Row+1, cursor.Col+1)

	used := lipgloss.Width(badge) + lipgloss.Width(name) + lipgloss.Width(validity) + lipgloss.Width(position)
	gap := strings.Repeat(" ", max(m.width-used, 1))
	return badge + m.theme.StatusBar.Render(name+gap) + validity + m.theme.StatusBar.Render(position)
}

func (m *Model) validity() string {
	if m.editor.Document().Type.Validator == nil {
		return ""
	}
	if m.validation.err == nil {
		return m.theme.Valid.Render("✓ valid")
	}
	var syntaxErr *filetype.SyntaxError
	if errors.As(m.validation.err, &syntaxErr) {
		return m.theme.Invalid.Render(fmt.Sprintf("✗ line %d", syntaxErr.Line))
	}
	return m.theme.Invalid.Render("✗ invalid")
}
