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
