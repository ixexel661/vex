package document

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/ixexel661/vex/internal/config"
	"github.com/ixexel661/vex/internal/filetype"
)

// schemaSet are the schemas a document is checked against. They are loaded
// when they are needed first, and again when one of their files changed.
type schemaSet struct {
	list     []config.Schema // as entered
	resolved []config.Schema // with absolute paths

	mu     sync.Mutex
	loaded bool
	stamps []stamp
	check  filetype.SchemaCheck
	err    error
}

// stamp tells whether a file changed.
type stamp struct {
	size, modified int64
}

func stampsOf(schemas []config.Schema) []stamp {
	stamps := make([]stamp, len(schemas))
	for i, s := range schemas {
		if info, err := os.Stat(s.Path); err == nil {
			stamps[i] = stamp{size: info.Size(), modified: info.ModTime().UnixNano()}
		}
	}
	return stamps
}

func (s *schemaSet) load(validator filetype.SchemaValidator) (filetype.SchemaCheck, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	stamps := stampsOf(s.resolved)
	if !s.loaded || !slices.Equal(stamps, s.stamps) {
		s.check, s.err = validator.Load(s.resolved)
		s.loaded, s.stamps = true, stamps
	}
	return s.check, s.err
}

// changed reports whether a file of the schemas changed since they were
// loaded.
func (s *schemaSet) changed() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.loaded && !slices.Equal(stampsOf(s.resolved), s.stamps)
}

// Schemas returns the schemas the document is checked against, as they
// were set.
func (d *Document) Schemas() []config.Schema {
	if d.schemas == nil {
		return nil
	}
	return slices.Clone(d.schemas.list)
}

// SetSchemas sets the schemas the document is checked against. Relative
// paths start in the directory of the document. The schemas are loaded
// with the next check, or by LoadSchemas.
func (d *Document) SetSchemas(schemas []config.Schema) {
	if len(schemas) == 0 || d.Type.Schemas == nil {
		d.schemas = nil
		return
	}
	resolved := slices.Clone(schemas)
	for i := range resolved {
		resolved[i].Path = d.Resolve(resolved[i].Path)
	}
	d.schemas = &schemaSet{list: slices.Clone(schemas), resolved: resolved}
}

// LoadSchemas loads the schemas now and returns why they cannot be loaded.
func (d *Document) LoadSchemas() error {
	if d.schemas == nil {
		return nil
	}
	_, err := d.schemas.load(d.Type.Schemas)
	return err
}

// SchemasChanged reports whether a file of the schemas changed since they
// were loaded.
func (d *Document) SchemasChanged() bool {
	return d.schemas != nil && d.schemas.changed()
}

// Resolve returns where path is: relative paths start in the directory of
// the document, and "~/" is the home directory.
func (d *Document) Resolve(path string) string {
	if rest, ok := cutHome(path); ok {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, rest)
		}
	}
	if filepath.IsAbs(path) {
		return path
	}
	dir := "."
	if d.Path != "" {
		dir = filepath.Dir(d.Path)
	}
	if abs, err := filepath.Abs(filepath.Join(dir, path)); err == nil {
		return abs
	}
	return path
}

func cutHome(path string) (string, bool) {
	if path == "~" {
		return "", true
	}
	for _, sep := range []string{"~/", "~" + string(filepath.Separator)} {
		if rest, ok := strings.CutPrefix(path, sep); ok {
			return rest, true
		}
	}
	return "", false
}
