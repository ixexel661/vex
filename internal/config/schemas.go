package config

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

const schemasHeader = "# The schemas vex checks files against. Press V in vex to change them.\n\n"

// Schema is a schema file, e.g. XSD, for the elements of a namespace. An
// empty Namespace is for elements without one.
type Schema struct {
	Namespace string `toml:"namespace"`
	Path      string `toml:"path"`
}

// LoadSchemas returns the schemas of the document at doc from the schemas
// file at path. A missing file has none.
func LoadSchemas(path, doc string) ([]Schema, error) {
	all, err := readSchemas(path)
	return all[doc], err
}

// SaveSchemas stores the schemas of the document at doc in the schemas file
// at path. The schemas of other documents stay, no schemas remove the entry
// of doc.
func SaveSchemas(path, doc string, schemas []Schema) error {
	all, err := readSchemas(path)
	if err != nil {
		return err
	}
	if len(schemas) == 0 {
		delete(all, doc)
	} else {
		all[doc] = schemas
	}
	var b bytes.Buffer
	b.WriteString(schemasHeader)
	enc := toml.NewEncoder(&b)
	enc.Indent = ""
	if err := enc.Encode(all); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, b.Bytes(), 0o644)
}

func readSchemas(path string) (map[string][]Schema, error) {
	all := map[string][]Schema{}
	_, err := toml.DecodeFile(path, &all)
	if errors.Is(err, fs.ErrNotExist) {
		return map[string][]Schema{}, nil
	}
	if err != nil {
		return nil, inFile(path, err)
	}
	return all, nil
}
