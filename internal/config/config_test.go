package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeConfig(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadMissingFileGivesDefaults(t *testing.T) {
	c, err := Load(filepath.Join(t.TempDir(), "config.toml"))
	if err != nil || c != Default() {
		t.Errorf("Load() = %+v, %v, want the defaults", c, err)
	}
}

func TestLoadKeepsDefaultsForMissingSettings(t *testing.T) {
	c, err := Load(writeConfig(t, "tab-width = 8\n[format]\nindent = \"tabs\"\n"))
	if err != nil {
		t.Fatal(err)
	}
	want := Default()
	want.TabWidth = 8
	want.Format.Indent = "\t"
	if c != want {
		t.Errorf("Load() = %+v, want %+v", c, want)
	}
}

func TestLoadReportsProblems(t *testing.T) {
	tests := []struct {
		name, content string
		want          func(c *Config)
		wantErr       string
	}{
		{
			name:    "unknown settings keep the others",
			content: "font = \"Hack\"\ntab-width = 8\n",
			want:    func(c *Config) { c.TabWidth = 8 },
			wantErr: "config.toml: unknown setting font",
		},
		{
			name:    "a value out of range gets its default",
			content: "tab-width = 0\n[format]\nself-close-empty = true\n",
			want:    func(c *Config) { c.Format.SelfCloseEmpty = true },
			wantErr: "config.toml: tab-width must be between 1 and 16, not 0",
		},
		{
			name:    "a broken file gives all defaults",
			content: "tab-width = 8\ntheme = \n",
			want:    func(c *Config) {},
			wantErr: "config.toml: line 2",
		},
		{
			name:    "an invalid indent gives all defaults",
			content: "tab-width = 8\n[format]\nindent = \"3 tabs\"\n",
			want:    func(c *Config) {},
			wantErr: `indent must be "tabs" or 1 to 8 spaces`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := Load(writeConfig(t, tt.content))
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error = %v, want it to contain %q", err, tt.wantErr)
			}
			want := Default()
			tt.want(&want)
			if c != want {
				t.Errorf("config = %+v, want %+v", c, want)
			}
		})
	}
}

func TestIndentText(t *testing.T) {
	for _, tt := range []struct {
		indent Indent
		text   string
	}{
		{"\t", "tabs"},
		{" ", "1 space"},
		{"    ", "4 spaces"},
	} {
		if got, _ := tt.indent.MarshalText(); string(got) != tt.text {
			t.Errorf("MarshalText(%q) = %q, want %q", tt.indent, got, tt.text)
		}
		var i Indent
		if err := i.UnmarshalText([]byte(tt.text)); err != nil || i != tt.indent {
			t.Errorf("UnmarshalText(%q) = %q, %v, want %q", tt.text, i, err, tt.indent)
		}
	}
	for _, text := range []string{"", "tab", "0 spaces", "9 spaces", "2", "2 tabs", "two spaces"} {
		var i Indent
		if err := i.UnmarshalText([]byte(text)); err == nil {
			t.Errorf("UnmarshalText(%q) = %q, want an error", text, i)
		}
	}
}

func TestSaveAndLoad(t *testing.T) {
	path := filepath.Join(t.TempDir(), "vex", "config.toml")
	c := Default()
	c.Theme = "dark"
	c.Format = Format{Indent: "\t", KeepBlankLines: true, WrapTagsLongerThan: 100}
	if err := Save(path, c); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(data), header) || !strings.Contains(string(data), "\nindent = \"tabs\"\n") {
		t.Errorf("file =\n%s\nwant the header and readable values", data)
	}
	if got, err := Load(path); err != nil || got != c {
		t.Errorf("Load() = %+v, %v, want %+v", got, err, c)
	}
}

func TestInstallKeepsAnExistingFile(t *testing.T) {
	path := writeConfig(t, "tab-width = 8\n")
	if err := Install(path); err != nil {
		t.Fatal(err)
	}
	if c, _ := Load(path); c.TabWidth != 8 {
		t.Error("Install overwrote the existing config")
	}

	missing := filepath.Join(t.TempDir(), "config.toml")
	if err := Install(missing); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(missing); err != nil {
		t.Errorf("Install did not create the config: %v", err)
	}
}
