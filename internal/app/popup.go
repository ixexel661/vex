package app

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// popup draws lines in a box with the title in its top border.
func (m *Model) popup(title string, lines []string) string {
	width := lipgloss.Width(title) + 2
	for _, line := range lines {
		width = max(width, lipgloss.Width(line))
	}
	border := m.theme.PopupBorder
	rows := []string{
		border.Render("╭─ ") + m.theme.PopupTitle.Render(title) + border.Render(" "+strings.Repeat("─", width-lipgloss.Width(title)-1)+"╮"),
	}
	for _, line := range lines {
		rows = append(rows, border.Render("│ ")+line+strings.Repeat(" ", width-lipgloss.Width(line))+border.Render(" │"))
	}
	rows = append(rows, border.Render("╰"+strings.Repeat("─", width+2)+"╯"))
	return strings.Join(rows, "\n")
}

// overlay draws box over the middle of the editor in content, the screen.
func (m *Model) overlay(content, box string) string {
	box = lipgloss.NewStyle().MaxWidth(m.width).MaxHeight(m.height).Render(box)
	x := max((m.width-lipgloss.Width(box))/2, 0)
	y := max((m.editorHeight()-lipgloss.Height(box))/2, 0)
	return lipgloss.NewCompositor(
		lipgloss.NewLayer(content),
		lipgloss.NewLayer(box).X(x).Y(y).Z(1),
	).Render()
}
