package theme

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// basic is the built-in theme every theme starts from.
const basic = "basic"

// Store finds themes by name, in a directory and among the themes built into
// vex. Every theme starts from the built-in basic theme, so a theme file only
// needs the styles it changes.
type Store struct {
	dir     string
	builtin fs.FS
}

func NewStore(dir string, builtin fs.FS) *Store {
	return &Store{dir: dir, builtin: builtin}
}

// Names returns the names of all themes, sorted.
func (s *Store) Names() []string {
	var names []string
	for _, fsys := range s.sources() {
		files, _ := fs.Glob(fsys, "*.toml")
		for _, f := range files {
			names = append(names, strings.TrimSuffix(f, ".toml"))
		}
	}
	slices.Sort(names)
	return slices.Compact(names)
}

// Load returns the theme with the given name. The theme is usable even with
// an error, e.g. it is basic if there is no theme with that name.
func (s *Store) Load(name string) (Theme, error) {
	base, err := read(s.builtin, basic, Theme{})
	if err != nil {
		return base, err
	}
	for _, fsys := range s.sources() {
		t, err := read(fsys, name, base)
		if !errors.Is(err, fs.ErrNotExist) {
			return t, err
		}
	}
	return base, fmt.Errorf("theme %q not found, using %s", name, basic)
}

// Install copies the built-in themes that are missing into the directory,
// where they are easy to find and copy.
func (s *Store) Install() error {
	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return err
	}
	files, err := fs.Glob(s.builtin, "*.toml")
	if err != nil {
		return err
	}
	for _, f := range files {
		path := filepath.Join(s.dir, f)
		if _, err := os.Stat(path); !errors.Is(err, fs.ErrNotExist) {
			if err != nil {
				return err
			}
			continue
		}
		data, err := fs.ReadFile(s.builtin, f)
		if err != nil {
			return err
		}
		if err := os.WriteFile(path, data, 0o644); err != nil {
			return err
		}
	}
	return nil
}

// sources are where themes are looked for, the first one wins.
func (s *Store) sources() []fs.FS {
	return []fs.FS{os.DirFS(s.dir), s.builtin}
}

func read(fsys fs.FS, name string, base Theme) (Theme, error) {
	file := name + ".toml"
	data, err := fs.ReadFile(fsys, file)
	if err != nil {
		return base, err
	}
	t, err := Parse(string(data), base)
	if err != nil {
		return t, fmt.Errorf("themes/%s: %s", file, strings.TrimPrefix(err.Error(), "toml: "))
	}
	return t, nil
}
