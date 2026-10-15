package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"charm.land/lipgloss/v2"
)

const settingsHints = "↑↓ select · ←→ change · enter save · esc cancel"

func (m *Model) settingsBox() string {
	s := m.settings
	labelWidth := 0
	for _, r := range s.rows {
		labelWidth = max(labelWidth, lipgloss.Width(r.label))
	}
	lines := make([]string, len(s.rows))
	for i, r := range s.rows {
		marker := " "
		if i == s.selected {
			marker = "▸"
		}
		lines[i] = fmt.Sprintf("%s %-*s  ‹ %s ›", marker, labelWidth, r.label, r.value(m.cfg))
	}
	footer := m.configLocation()
	width := lipgloss.Width(footer)
	for _, line := range lines {
		width = max(width, lipgloss.Width(line))
	}
	lines[s.selected] = m.theme.Selected.Render(lines[s.selected] + strings.Repeat(" ", width-lipgloss.Width(lines[s.selected])))
	return m.popup("Settings", append(lines, "", m.theme.Hint.Render(footer)))
}

// configLocation is the path of the config file, with ~ for the home
// directory.
func (m *Model) configLocation() string {
	if home, err := os.UserHomeDir(); err == nil {
		if rest, ok := strings.CutPrefix(m.configPath, home+string(filepath.Separator)); ok {
			return filepath.Join("~", rest)
		}
	}
	return m.configPath
}
