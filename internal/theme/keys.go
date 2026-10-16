package theme

import (
	"charm.land/lipgloss/v2"

	"github.com/ixexel661/vex/internal/highlight"
)

// kinds maps the keys of the [syntax] table of a theme file to the kinds of
// text they style.
var kinds = map[string]highlight.Kind{
	"plain":         highlight.Plain,
	"punctuation":   highlight.Punctuation,
	"tag":           highlight.Tag,
	"attribute":     highlight.Attribute,
	"string":        highlight.String,
	"comment":       highlight.Comment,
	"meta":          highlight.Meta,
	"entity":        highlight.Entity,
	"cdata":         highlight.CData,
	"match":         highlight.Match,
	"current-match": highlight.CurrentMatch,
	"linked-tag":    highlight.LinkedTag,
}

// ui maps the keys of the [ui] table of a theme file to the styles of t.
func (t *Theme) ui() map[string]*lipgloss.Style {
	return map[string]*lipgloss.Style{
		"line-number":         &t.LineNumber,
		"current-line-number": &t.CurrentLineNumber,
		"filler":              &t.Filler,
		"status-bar":          &t.StatusBar,
		"normal-mode":         &t.NormalMode,
		"insert-mode":         &t.InsertMode,
		"valid":               &t.Valid,
		"invalid":             &t.Invalid,
		"message":             &t.Message,
		"error":               &t.Error,
		"question":            &t.Question,
		"hint":                &t.Hint,
		"search-mode":         &t.SearchMode,
		"selected":            &t.Selected,
		"matched-rune":        &t.MatchedRune,
		"popup-border":        &t.PopupBorder,
		"popup-title":         &t.PopupTitle,
	}
}
