package filetype

import (
	"path/filepath"
	"slices"
	"strings"
)

type Registry struct {
	types []Type
}

func NewRegistry(types ...Type) *Registry {
	return &Registry{types: types}
}

func (r *Registry) Register(t Type) {
	r.types = append(r.types, t)
}

// Default is the type of new files without a name or an extension: the
// first one registered.
func (r *Registry) Default() Type {
	if len(r.types) == 0 {
		return Plain()
	}
	return r.types[0]
}

// Detect picks the file type by the extension of path.
func (r *Registry) Detect(path string) Type {
	ext := strings.ToLower(filepath.Ext(path))
	for _, t := range r.types {
		if slices.Contains(t.Extensions, ext) {
			return t
		}
	}
	return Plain()
}
