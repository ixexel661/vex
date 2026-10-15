package theme

import (
	"charm.land/lipgloss/v2"

	"github.com/ixexel661/vex/internal/highlight"
)

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
}

// Default uses the basic ANSI colors, so it follows the terminal's own
// color scheme and works on light and dark backgrounds.
func Default() Theme {
	plain := lipgloss.NewStyle()
	bar := plain.Background(lipgloss.BrightBlack).Foreground(lipgloss.BrightWhite)
	badge := plain.Foreground(lipgloss.Black).Bold(true).Padding(0, 1)

	return Theme{
		Syntax: map[highlight.Kind]lipgloss.Style{
			highlight.Punctuation:  plain.Foreground(lipgloss.BrightBlack),
			highlight.Tag:          plain.Foreground(lipgloss.Blue),
			highlight.Attribute:    plain.Foreground(lipgloss.Cyan),
			highlight.String:       plain.Foreground(lipgloss.Green),
			highlight.Comment:      plain.Foreground(lipgloss.BrightBlack).Italic(true),
			highlight.Meta:         plain.Foreground(lipgloss.Magenta),
			highlight.Entity:       plain.Foreground(lipgloss.Yellow),
			highlight.CData:        plain.Foreground(lipgloss.Yellow),
			highlight.Match:        plain.Foreground(lipgloss.Black).Background(lipgloss.Yellow),
			highlight.CurrentMatch: plain.Foreground(lipgloss.BrightWhite).Background(lipgloss.Magenta).Bold(true),
		},

		LineNumber:        plain.Foreground(lipgloss.BrightBlack),
		CurrentLineNumber: plain.Foreground(lipgloss.Yellow),
		Filler:            plain.Foreground(lipgloss.BrightBlack),

		StatusBar:  bar,
		NormalMode: badge.Background(lipgloss.Blue),
		InsertMode: badge.Background(lipgloss.Green),
		Valid:      bar.Foreground(lipgloss.BrightGreen),
		Invalid:    bar.Foreground(lipgloss.BrightRed).Bold(true),

		Message:  plain,
		Error:    plain.Foreground(lipgloss.Red),
		Question: plain.Foreground(lipgloss.Yellow).Bold(true),
		Hint:     plain.Foreground(lipgloss.BrightBlack),

		SearchMode:  badge.Background(lipgloss.Magenta),
		Selected:    bar,
		MatchedRune: plain.Foreground(lipgloss.Yellow).Bold(true),
	}
}

// Token returns the style for a highlighted piece of text. Kinds without an
// entry, like plain text, stay unstyled.
func (t Theme) Token(kind highlight.Kind) lipgloss.Style {
	return t.Syntax[kind]
}
