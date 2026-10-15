package theme

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"testing/fstest"

	"charm.land/lipgloss/v2"
)

var builtin = fstest.MapFS{
	"basic.toml": {Data: []byte("[ui]\nhint = { fg = \"blue\" }\nerror = { fg = \"red\" }\n")},
	"light.toml": {Data: []byte("[ui]\nhint = { fg = \"black\" }\n")},
}

// newStore returns a store whose directory holds the given files.
func newStore(t *testing.T, files map[string]string) *Store {
	t.Helper()
	dir := t.TempDir()
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return NewStore(dir, builtin)
}

func fg(c string) lipgloss.Style {
	color, _ := parseColor(c)
	return lipgloss.NewStyle().Foreground(color)
}

func TestStoreNames(t *testing.T) {
	s := newStore(t, map[string]string{"dark.toml": "", "basic.toml": "", "notes.txt": ""})
	if got, want := s.Names(), []string{"basic", "dark", "light"}; !slices.Equal(got, want) {
		t.Errorf("Names() = %v, want %v", got, want)
	}
}

func TestStoreLoad(t *testing.T) {
	s := newStore(t, map[string]string{
		"dark.toml":   "[ui]\nhint = { fg = \"green\" }\n",
		"light.toml":  "[ui]\nhint = { fg = \"white\" }\n",
		"broken.toml": "[ui\n",
	})
	tests := []struct {
		name, theme string
		wantHint    lipgloss.Style
		wantErr     string
	}{
		{"from the directory", "dark", fg("green"), ""},
		{"the directory comes first", "light", fg("white"), ""},
		{"built in", "basic", fg("blue"), ""},
		{"missing", "solarized", fg("blue"), `theme "solarized" not found, using basic`},
		{"broken", "broken", fg("blue"), "themes/broken.toml: line 2"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			th, err := s.Load(tt.theme)
			if tt.wantErr == "" && err != nil || tt.wantErr != "" && (err == nil || !strings.Contains(err.Error(), tt.wantErr)) {
				t.Errorf("error = %v, want %q", err, tt.wantErr)
			}
			if !same(th.Hint, tt.wantHint) {
				t.Errorf("hint renders %q, want %q", th.Hint.Render("x"), tt.wantHint.Render("x"))
			}
			if !same(th.Error, fg("red")) {
				t.Error("the styles the theme leaves out do not come from basic")
			}
		})
	}
}

func TestStoreInstall(t *testing.T) {
	s := newStore(t, map[string]string{"basic.toml": "# mine\n"})
	if err := s.Install(); err != nil {
		t.Fatal(err)
	}
	for name, want := range map[string]string{
		"basic.toml": "# mine\n",
		"light.toml": string(builtin["light.toml"].Data),
	} {
		got, err := os.ReadFile(filepath.Join(s.dir, name))
		if err != nil || string(got) != want {
			t.Errorf("%s = %q, %v, want %q", name, got, err, want)
		}
	}
}
