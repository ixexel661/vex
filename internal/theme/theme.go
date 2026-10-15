package theme

import (
	"charm.land/lipgloss/v2"

	"github.com/ixexel661/vex/internal/highlight"
)

// Theme holds the styles vex draws with. They come from theme files, see
// Store and themes/basic.toml.
type Theme struct {
	Syntax map[highlight.Kind]lipgloss.Style

	LineNumber        lipgloss.Style
	CurrentLineNumber lipgloss.Style
	Filler            lipgloss.Style

	StatusBar  lipgloss.Style
	NormalMode lipgloss.Style
	InsertMode lipgloss.Style
	Valid      lipgloss.Style
	Invalid    lipgloss.Style

	Message  lipgloss.Style
	Error    lipgloss.Style
	Question lipgloss.Style
	Hint     lipgloss.Style

	SearchMode  lipgloss.Style
	Selected    lipgloss.Style
	MatchedRune lipgloss.Style

	PopupBorder lipgloss.Style
	PopupTitle  lipgloss.Style
}

// Token returns the style for a highlighted piece of text. Kinds without an
// entry stay unstyled.
func (t Theme) Token(kind highlight.Kind) lipgloss.Style {
	return t.Syntax[kind]
}
