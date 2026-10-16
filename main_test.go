package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ixexel661/vex/internal/filetype"
	"github.com/ixexel661/vex/internal/filetype/xml"
)

func fileTypes() *filetype.Registry {
	return filetype.NewRegistry(xml.Type())
}

func TestCreateDocument(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test")
	var out strings.Builder
	doc, err := createDocument(path, fileTypes(), strings.NewReader(""), &out)
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path + ".xml")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != xml.Type().Template || doc.Modified() || doc.Type.Name != "XML" {
		t.Errorf("file = %q, modified = %v, type = %s, want the saved XML template", data, doc.Modified(), doc.Type.Name)
	}
	if out.Len() != 0 {
		t.Errorf("asked %q for a file that did not exist", out.String())
	}
}

func TestCreateDocumentAsksBeforeOverwriting(t *testing.T) {
	tests := []struct {
		answer    string
		overwrite bool
	}{
		{"y\n", true},
		{"Yes\n", true},
		{"n\n", false},
		{"\n", false},
		{"", false},
	}
	for _, tt := range tests {
		path := filepath.Join(t.TempDir(), "test.xml")
		if err := os.WriteFile(path, []byte("old"), 0o644); err != nil {
			t.Fatal(err)
		}
		var out strings.Builder
		_, err := createDocument(path, fileTypes(), strings.NewReader(tt.answer), &out)
		if want := path + " exists already. Overwrite it? [y/N] "; out.String() != want {
			t.Errorf("question = %q, want %q", out.String(), want)
		}
		data, _ := os.ReadFile(path)
		if tt.overwrite && (err != nil || string(data) != xml.Type().Template) {
			t.Errorf("answer %q: err = %v, file = %q, want it overwritten", tt.answer, err, data)
		}
		if !tt.overwrite && (err == nil || string(data) != "old") {
			t.Errorf("answer %q: err = %v, file = %q, want an error and the old file", tt.answer, err, data)
		}
	}
}

func TestOpenDocument(t *testing.T) {
	doc, isNew, err := openDocument([]string{"new"}, fileTypes())
	if err != nil || !isNew || doc.Path != "" || doc.Type.Name != "XML" {
		t.Errorf("vex new: err = %v, new = %v, path = %q, want a new unnamed XML document", err, isNew, doc.Path)
	}

	path := filepath.Join(t.TempDir(), "test.xml")
	doc, isNew, err = openDocument([]string{path}, fileTypes())
	if err != nil || isNew || doc.Path != path {
		t.Errorf("vex file: err = %v, new = %v, path = %q, want the document of the file", err, isNew, doc.Path)
	}

	for _, args := range [][]string{nil, {"a.xml", "b.xml"}, {"new", "a.xml", "b.xml"}} {
		if _, _, err := openDocument(args, fileTypes()); !errors.Is(err, errUsage) {
			t.Errorf("args %q: err = %v, want the usage", args, err)
		}
	}
}

func TestVersionOf(t *testing.T) {
	tests := []struct {
		tagged, module, want string
	}{
		{"v1.2.0", "v0.0.0-20261015123645-882f7945108b+dirty", "v1.2.0"},
		{"", "v1.2.0", "v1.2.0"},
		{"", "v1.3.0-beta.1", "v1.3.0-beta.1"},
		{"", "", "dev"},
		{"", "(devel)", "dev"},
		{"", "v0.0.0-20261015123645-882f7945108b+dirty", "dev"},
		{"", "v1.2.1-0.20261015123645-882f7945108b", "dev"},
	}
	for _, tt := range tests {
		if got := versionOf(tt.tagged, tt.module); got != tt.want {
			t.Errorf("versionOf(%q, %q) = %q, want %q", tt.tagged, tt.module, got, tt.want)
		}
	}
}
