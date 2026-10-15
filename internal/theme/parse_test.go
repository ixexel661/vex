package theme

import (
	"image/color"
	"os"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/BurntSushi/toml"

	"github.com/ixexel661/vex/internal/highlight"
)

func TestParseColor(t *testing.T) {
	for s, want := range map[string]color.Color{
		"black":        lipgloss.Black,
		"yellow":       lipgloss.Yellow,
		"bright-black": lipgloss.BrightBlack,
		"bright-white": lipgloss.BrightWhite,
		"3":            lipgloss.Yellow,
		"208":          lipgloss.Color("208"),
		"#ff8700":      lipgloss.Color("#ff8700"),
		"#F80":         lipgloss.Color("#F80"),
	} {
		if got, err := parseColor(s); err != nil || got != want {
			t.Errorf("parseColor(%q) = %v, %v, want %v", s, got, err, want)
		}
	}
	for _, s := range []string{"yelow", "Yellow", "256", "-1", "#ff87", "ff8700", "#gg8700"} {
		if got, err := parseColor(s); err == nil {
			t.Errorf("parseColor(%q) = %v, want an error", s, got)
		}
	}
}

// same reports whether two styles look the same.
func same(a, b lipgloss.Style) bool {
	return a.Render("x") == b.Render("x")
}

func TestParseStartsFromBase(t *testing.T) {
	plain := lipgloss.NewStyle()
	base := Theme{
		Hint:   plain.Bold(true),
		Syntax: map[highlight.Kind]lipgloss.Style{highlight.Tag: plain.Underline(true)},
	}
	th, err := Parse("[ui]\nerror = { fg = \"red\", bg = \"#000000\" }\n[syntax]\ncomment = { italic = true }\n", base)
	if err != nil {
		t.Fatal(err)
	}
	checks := []struct {
		name      string
		got, want lipgloss.Style
	}{
		{"ui.error", th.Error, plain.Foreground(lipgloss.Red).Background(lipgloss.Color("#000000"))},
		{"syntax.comment", th.Token(highlight.Comment), plain.Italic(true)},
		{"ui.hint from base", th.Hint, base.Hint},
		{"syntax.tag from base", th.Token(highlight.Tag), base.Syntax[highlight.Tag]},
	}
	for _, c := range checks {
		if !same(c.got, c.want) {
			t.Errorf("%s renders %q, want %q", c.name, c.got.Render("x"), c.want.Render("x"))
		}
	}
	if _, ok := base.Syntax[highlight.Comment]; ok {
		t.Error("Parse changed the syntax styles of base")
	}
}

func TestParseReportsWhatItSkips(t *testing.T) {
	data := `
[ui]
hint = { fg = "blue" }
error = { fg = "redd" }
blink = {}

[syntax]
tag = { forground = "red" }

[colors]
red = {}
`
	th, err := Parse(data, Theme{})
	if !same(th.Hint, lipgloss.NewStyle().Foreground(lipgloss.Blue)) {
		t.Error("the valid style was not applied")
	}
	if err == nil {
		t.Fatal("no error")
	}
	for _, want := range []string{
		`ui.error: unknown color "redd"`,
		"ui.blink: unknown style",
		"unknown key syntax.tag.forground",
		"colors.red: unknown table",
	} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q does not mention %q", err, want)
		}
	}
}

func TestBasicThemeIsComplete(t *testing.T) {
	data, err := os.ReadFile("../../themes/basic.toml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Parse(string(data), Theme{}); err != nil {
		t.Errorf("basic.toml: %v", err)
	}
	var file map[string]map[string]spec
	if _, err := toml.Decode(string(data), &file); err != nil {
		t.Fatal(err)
	}
	for name := range kinds {
		if _, ok := file["syntax"][name]; !ok {
			t.Errorf("basic.toml has no style syntax.%s", name)
		}
	}
	for name := range (&Theme{}).ui() {
		if _, ok := file["ui"][name]; !ok {
			t.Errorf("basic.toml has no style ui.%s", name)
		}
	}
}
