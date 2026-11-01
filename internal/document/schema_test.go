package document

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/ixexel661/vex/internal/config"
	"github.com/ixexel661/vex/internal/filetype"
)

// fakeSchemas finds a problem in texts with "bad" in them. A schema file
// with "broken" in it does not load.
type fakeSchemas struct {
	loaded *[][]config.Schema
}

func (fakeSchemas) Name() string { return "Fake" }

func (f fakeSchemas) Load(schemas []config.Schema) (filetype.SchemaCheck, error) {
	*f.loaded = append(*f.loaded, schemas)
	for _, s := range schemas {
		data, err := os.ReadFile(s.Path)
		if err != nil {
			return nil, err
		}
		if strings.Contains(string(data), "broken") {
			return nil, errors.New("broken schema")
		}
	}
	return func(src string) []filetype.Problem {
		if strings.Contains(src, "bad") {
			return []filetype.Problem{{Message: "bad"}}
		}
		return nil
	}, nil
}

func (fakeSchemas) Named(string) []config.Schema { return nil }

// schemaDoc opens text in a temporary directory with a schema file next to
// it and returns the loads of the schemas.
func schemaDoc(t *testing.T, text string, validator filetype.Validator) (*Document, *[][]config.Schema) {
	t.Helper()
	loaded := &[][]config.Schema{}
	path := writeFile(t, text)
	writeSchema(t, path, "ok")
	doc, err := Open(path, filetype.Type{Name: "Fake", Validator: validator, Schemas: fakeSchemas{loaded}})
	if err != nil {
		t.Fatal(err)
	}
	doc.SetSchemas([]config.Schema{{Namespace: "urn:a", Path: "s.xsd"}})
	return doc, loaded
}

func writeSchema(t *testing.T, doc, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(filepath.Dir(doc), "s.xsd"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestValidationAgainstSchemas(t *testing.T) {
	doc, loaded := schemaDoc(t, "<bad/>", nil)
	v := doc.Validation()()
	if !v.Schema || v.SchemaErr != nil || len(v.Problems) != 1 {
		t.Fatalf("Validation() = %+v, want the problem of the schema", v)
	}
	want := []config.Schema{{Namespace: "urn:a", Path: filepath.Join(filepath.Dir(doc.Path), "s.xsd")}}
	if len(*loaded) != 1 || !slices.Equal((*loaded)[0], want) {
		t.Errorf("loaded %v, want %v once", *loaded, want)
	}
	if got := doc.Schemas(); !slices.Equal(got, []config.Schema{{Namespace: "urn:a", Path: "s.xsd"}}) {
		t.Errorf("Schemas() = %v, want them as they were set", got)
	}
}

func TestSchemasAreNotCheckedInTextThatIsNotWellFormed(t *testing.T) {
	doc, _ := schemaDoc(t, "<bad/>", rejectAll{})
	v := doc.Validation()()
	if v.Err == nil || !v.Schema || v.Problems != nil {
		t.Errorf("Validation() = %+v, want only the syntax error", v)
	}
}

func TestSchemaThatDoesNotLoad(t *testing.T) {
	doc, _ := schemaDoc(t, "<bad/>", nil)
	writeSchema(t, doc.Path, "broken")
	if err := doc.LoadSchemas(); err == nil {
		t.Error("LoadSchemas() = nil, want the error")
	}
	if v := doc.Validation()(); v.SchemaErr == nil || v.Problems != nil {
		t.Errorf("Validation() = %+v, want the error of the schema", v)
	}
}

func TestChangedSchemaIsLoadedAgain(t *testing.T) {
	doc, loaded := schemaDoc(t, "<bad/>", nil)
	if doc.SchemasChanged() {
		t.Error("the schemas changed before they were loaded")
	}
	doc.Validation()()
	doc.Validation()()
	if len(*loaded) != 1 || doc.SchemasChanged() {
		t.Fatalf("loaded %d times, want once", len(*loaded))
	}
	writeSchema(t, doc.Path, "broken, and longer")
	if !doc.SchemasChanged() {
		t.Error("SchemasChanged() = false after the file changed")
	}
	if v := doc.Validation()(); v.SchemaErr == nil || len(*loaded) != 2 {
		t.Errorf("after the change: %+v, loaded %d times, want the new schema", v, len(*loaded))
	}
}

func TestNoSchemas(t *testing.T) {
	doc, _ := schemaDoc(t, "<bad/>", nil)
	doc.SetSchemas(nil)
	if v := doc.Validation()(); v.Schema || v.Problems != nil || doc.Schemas() != nil {
		t.Errorf("Validation() = %+v without schemas", v)
	}
	plain := New("", filetype.Plain())
	plain.SetSchemas([]config.Schema{{Path: "a.xsd"}})
	if plain.Schemas() != nil {
		t.Error("a file type without schemas has some")
	}
}

func TestResolve(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip(err)
	}
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	doc := New(filepath.Join(dir, "a.xml"), filetype.Plain())
	untitled := New("", filetype.Plain())
	tests := []struct {
		doc        *Document
		path, want string
	}{
		{doc, "s.xsd", filepath.Join(dir, "s.xsd")},
		{doc, "../s.xsd", filepath.Join(filepath.Dir(dir), "s.xsd")},
		{doc, "~/x/s.xsd", filepath.Join(home, "x", "s.xsd")},
		{doc, filepath.Join(wd, "s.xsd"), filepath.Join(wd, "s.xsd")},
		{untitled, "s.xsd", filepath.Join(wd, "s.xsd")},
	}
	for _, tt := range tests {
		if got := tt.doc.Resolve(tt.path); got != tt.want {
			t.Errorf("Resolve(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}
