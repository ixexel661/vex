package config

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestSaveAndLoadSchemas(t *testing.T) {
	path := filepath.Join(t.TempDir(), "vex", "schemas.toml")
	if got, err := LoadSchemas(path, "/a.xml"); err != nil || got != nil {
		t.Fatalf("without a file: %v, %v, want none", got, err)
	}

	a := []Schema{{Namespace: "urn:a", Path: "a.xsd"}, {Path: "/x/plain.xsd"}}
	b := []Schema{{Namespace: "urn:b", Path: "b.xsd"}}
	if err := SaveSchemas(path, "/a.xml", a); err != nil {
		t.Fatal(err)
	}
	if err := SaveSchemas(path, `C:\b "1".xml`, b); err != nil {
		t.Fatal(err)
	}
	for doc, want := range map[string][]Schema{"/a.xml": a, `C:\b "1".xml`: b} {
		if got, err := LoadSchemas(path, doc); err != nil || !slices.Equal(got, want) {
			t.Errorf("LoadSchemas(%q) = %v, %v, want %v", doc, got, err, want)
		}
	}

	if err := SaveSchemas(path, "/a.xml", nil); err != nil {
		t.Fatal(err)
	}
	if got, _ := LoadSchemas(path, "/a.xml"); got != nil {
		t.Errorf("after removing them, the schemas are %v", got)
	}
	if got, _ := LoadSchemas(path, `C:\b "1".xml`); !slices.Equal(got, b) {
		t.Errorf("the schemas of the other file are %v, want %v", got, b)
	}
}

func TestBrokenSchemasFileStays(t *testing.T) {
	path := filepath.Join(t.TempDir(), "schemas.toml")
	if err := os.WriteFile(path, []byte("broken ="), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadSchemas(path, "/a.xml"); err == nil || !strings.HasPrefix(err.Error(), "schemas.toml: ") {
		t.Errorf("LoadSchemas() error = %v, want one about the file", err)
	}
	if err := SaveSchemas(path, "/a.xml", []Schema{{Path: "a.xsd"}}); err == nil {
		t.Error("SaveSchemas overwrote a file it cannot read")
	}
}
