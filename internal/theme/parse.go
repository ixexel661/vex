package theme

import (
	"errors"
	"fmt"
	"image/color"
	"maps"
	"regexp"
	"slices"
	"strconv"

	"charm.land/lipgloss/v2"
	"github.com/BurntSushi/toml"

	"github.com/ixexel661/vex/internal/highlight"
)

// spec is a style as a theme file writes it, e.g. { fg = "yellow", bold = true }.
type spec struct {
	Fg        string `toml:"fg"`
	Bg        string `toml:"bg"`
	Bold      bool   `toml:"bold"`
	Italic    bool   `toml:"italic"`
	Underline bool   `toml:"underline"`
}

// colorNames are the 16 basic terminal colors, in the order of their numbers.
var colorNames = []string{
	"black", "red", "green", "yellow", "blue", "magenta", "cyan", "white",
	"bright-black", "bright-red", "bright-green", "bright-yellow",
	"bright-blue", "bright-magenta", "bright-cyan", "bright-white",
}

var hexColor = regexp.MustCompile(`^#([0-9a-fA-F]{3}){1,2}$`)

// Parse reads a theme file on top of base: styles the file leaves out keep
// the style of base. The theme is usable even with an error, which names the
// styles that were skipped.
func Parse(data string, base Theme) (Theme, error) {
	var file map[string]map[string]spec
	md, err := toml.Decode(data, &file)
	if err != nil {
		return base, err
	}
	t := base
	t.Syntax = maps.Clone(base.Syntax)
	if t.Syntax == nil {
		t.Syntax = map[highlight.Kind]lipgloss.Style{}
	}
	var errs []error
	for _, key := range md.Undecoded() {
		errs = append(errs, fmt.Errorf("unknown key %s", key))
	}
	for _, table := range slices.Sorted(maps.Keys(file)) {
		for _, name := range slices.Sorted(maps.Keys(file[table])) {
			if err := t.set(table, name, file[table][name]); err != nil {
				errs = append(errs, fmt.Errorf("%s.%s: %w", table, name, err))
			}
		}
	}
	return t, errors.Join(errs...)
}

func (t *Theme) set(table, name string, s spec) error {
	style, err := s.style()
	if err != nil {
		return err
	}
	switch table {
	case "syntax":
		kind, ok := kinds[name]
		if !ok {
			return errors.New("unknown style")
		}
		t.Syntax[kind] = style
	case "ui":
		field, ok := t.ui()[name]
		if !ok {
			return errors.New("unknown style")
		}
		*field = style
	default:
		return errors.New("unknown table")
	}
	return nil
}

func (s spec) style() (lipgloss.Style, error) {
	style := lipgloss.NewStyle()
	if s.Fg != "" {
		c, err := parseColor(s.Fg)
		if err != nil {
			return style, err
		}
		style = style.Foreground(c)
	}
	if s.Bg != "" {
		c, err := parseColor(s.Bg)
		if err != nil {
			return style, err
		}
		style = style.Background(c)
	}
	if s.Bold {
		style = style.Bold(true)
	}
	if s.Italic {
		style = style.Italic(true)
	}
	if s.Underline {
		style = style.Underline(true)
	}
	return style, nil
}

// parseColor reads a color name, a number from 0 to 255 or a hex color like
// "#ff8700".
func parseColor(s string) (color.Color, error) {
	if i := slices.Index(colorNames, s); i >= 0 {
		return lipgloss.Color(strconv.Itoa(i)), nil
	}
	if n, err := strconv.Atoi(s); err == nil && n >= 0 && n <= 255 {
		return lipgloss.Color(s), nil
	}
	if hexColor.MatchString(s) {
		return lipgloss.Color(s), nil
	}
	return nil, fmt.Errorf("unknown color %q", s)
}
