package app

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/ixexel661/vex/internal/mode"
)

func (m *Model) View() tea.View {
	v := tea.NewView("")
	v.AltScreen = true
	v.WindowTitle = "vex – " + m.editor.Document().Name()
	if m.width == 0 {
		return v
	}

	line := lipgloss.NewStyle().MaxWidth(m.width)
	content := m.editor.View()
	if height := m.listHeight(); height > 0 {
		content += "\n" + line.Render(m.matchList(height))
	}
	content += "\n" + line.Render(m.statusBar()) + "\n" + line.Render(m.messageLine())
	if m.settings != nil {
		content = m.overlay(content, m.settingsBox())
	}
	v.SetContent(content)
	v.Cursor = m.cursor()
	return v
}

func (m *Model) cursor() *tea.Cursor {
	if m.settings != nil {
		return nil
	}
	if m.searching {
		c := tea.NewCursor(min(lipgloss.Width(m.searchInput()), m.width-1), m.height-1)
		c.Shape = tea.CursorBar
		return c
	}
	x, y := m.editor.CursorPosition()
	c := tea.NewCursor(x, y)
	if m.mode == mode.Insert {
		c.Shape = tea.CursorBar
	}
	return c
}
