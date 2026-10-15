package app

import (
	"fmt"
	"path/filepath"
	"slices"
	"strconv"

	"github.com/ixexel661/vex/internal/config"
	"github.com/ixexel661/vex/internal/theme"
)

// settingsPopup is the open settings popup. Changes apply at once, cancel
// goes back to the config and theme from when it opened.
type settingsPopup struct {
	rows     []setting
	selected int
	cfg      config.Config
	theme    theme.Theme
}

// setting is a row of the settings popup.
type setting struct {
	label  string
	value  func(c config.Config) string
	change func(c *config.Config, step int)
}

func settingRows(themes []string) []setting {
	return []setting{
		option("Theme", themes, func(name string) string { return name },
			func(c *config.Config) *string { return &c.Theme }),
		option("Tab width", []int{2, 4, 8}, strconv.Itoa,
			func(c *config.Config) *int { return &c.TabWidth }),
		option("Indent", []config.Indent{"\t", "  ", "    "}, config.Indent.String,
			func(c *config.Config) *config.Indent { return &c.Format.Indent }),
		option("Empty elements", []bool{false, true}, either("<a/>", "<a></a>"),
			func(c *config.Config) *bool { return &c.Format.SelfCloseEmpty }),
		option("Blank lines", []bool{false, true}, either("keep", "remove"),
			func(c *config.Config) *bool { return &c.Format.KeepBlankLines }),
		option("Wrap tags", []int{0, 80, 100, 120}, wrapLimit,
			func(c *config.Config) *int { return &c.Format.WrapTagsLongerThan }),
	}
}

// option is a setting that steps through choices. field points to the value
// in a config, show turns it into text.
func option[T comparable](label string, choices []T, show func(T) string, field func(c *config.Config) *T) setting {
	return setting{
		label: label,
		value: func(c config.Config) string {
			return show(*field(&c))
		},
		change: func(c *config.Config, step int) {
			v := field(c)
			*v = cycle(choices, *v, step)
		},
	}
}

// cycle returns the choice step places away from current and wraps around.
// A current value that is no choice, e.g. typed into the config file, steps
// to the first or last choice.
func cycle[T comparable](choices []T, current T, step int) T {
	if len(choices) == 0 {
		return current
	}
	i := slices.Index(choices, current)
	if i < 0 && step < 0 {
		i = 0
	}
	n := len(choices)
	return choices[((i+step)%n+n)%n]
}

func either(yes, no string) func(bool) string {
	return func(b bool) string {
		if b {
			return yes
		}
		return no
	}
}

func wrapLimit(n int) string {
	if n == 0 {
		return "never"
	}
	return fmt.Sprintf("over %d", n)
}

func (m *Model) OpenSettings() {
	m.settings = &settingsPopup{
		rows:  settingRows(m.themes.Names()),
		cfg:   m.cfg,
		theme: m.theme,
	}
}

func (m *Model) settingsKey(key string) {
	s := m.settings
	switch key {
	case "up":
		s.selected = (s.selected + len(s.rows) - 1) % len(s.rows)
	case "down":
		s.selected = (s.selected + 1) % len(s.rows)
	case "left":
		m.changeSetting(-1)
	case "right":
		m.changeSetting(1)
	case "enter":
		m.saveSettings()
	case "esc", "ctrl+c":
		m.cancelSettings()
	}
}

func (m *Model) changeSetting(step int) {
	cfg := m.cfg
	m.settings.rows[m.settings.selected].change(&cfg, step)
	m.setConfig(cfg)
}

// setConfig applies cfg at once. The theme is only loaded if its name
// changes.
func (m *Model) setConfig(cfg config.Config) {
	if cfg.Theme != m.cfg.Theme {
		th, err := m.themes.Load(cfg.Theme)
		m.reportError(err)
		m.theme = th
	}
	m.cfg = cfg
	m.editor.Configure(m.cfg, m.theme)
}

func (m *Model) saveSettings() {
	if err := config.Save(m.configPath, m.cfg); err != nil {
		m.reportError(fmt.Errorf("cannot save the settings: %w", err))
		return
	}
	m.settings = nil
	m.Notify("Saved " + filepath.Base(m.configPath))
}

func (m *Model) cancelSettings() {
	m.cfg, m.theme = m.settings.cfg, m.settings.theme
	m.editor.Configure(m.cfg, m.theme)
	m.settings = nil
}
