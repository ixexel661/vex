package document

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/ixexel661/vex/internal/buffer"
	"github.com/ixexel661/vex/internal/filetype"
)

func writeFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.xml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestSaveKeepsLineEndingsAndBOM(t *testing.T) {
	original := "\uFEFF<a>\r\n  <b/>\r\n</a>\r\n"
	path := writeFile(t, original)

	doc, err := Open(path, filetype.Plain())
	if err != nil {
		t.Fatal(err)
	}
	if got := doc.Buffer.Text(); got != "<a>\n  <b/>\n</a>\n" {
		t.Errorf("buffer text = %q", got)
	}
	if err := doc.Save(); err != nil {
		t.Fatal(err)
	}
	if got := readFile(t, path); got != original {
		t.Errorf("saved file = %q, want %q", got, original)
	}
}

func TestOpenMissingFileCreatesItOnSave(t *testing.T) {
	path := filepath.Join(t.TempDir(), "new.xml")
	doc, err := Open(path, filetype.Plain())
	if err != nil {
		t.Fatal(err)
	}
	doc.Buffer.Insert(buffer.Position{}, "<a/>")
	if err := doc.Save(); err != nil {
		t.Fatal(err)
	}
	if got := readFile(t, path); got != "<a/>" {
		t.Errorf("saved file = %q, want %q", got, "<a/>")
	}
}

func TestOpenRejectsNonUTF8(t *testing.T) {
	path := writeFile(t, "<a>\xff</a>")
	if _, err := Open(path, filetype.Plain()); err == nil {
		t.Error("Open() succeeded, want an error")
	}
}

func TestModified(t *testing.T) {
	doc, err := Open(writeFile(t, "<a/>"), filetype.Plain())
	if err != nil {
		t.Fatal(err)
	}
	if doc.Modified() {
		t.Error("new document is modified")
	}
	doc.Buffer.Insert(buffer.Position{}, "x")
	if !doc.Modified() {
		t.Error("edited document is not modified")
	}
	if err := doc.Save(); err != nil {
		t.Fatal(err)
	}
	if doc.Modified() {
		t.Error("saved document is modified")
	}
}

type rejectAll struct{}

func (rejectAll) Validate(string) error { return errors.New("invalid") }

func TestValidate(t *testing.T) {
	fileType := filetype.Type{Name: "Strict", Validator: rejectAll{}}

	doc, err := Open(writeFile(t, "<a/>"), fileType)
	if err != nil {
		t.Fatal(err)
	}
	if doc.Validate() == nil {
		t.Error("Validate() = nil, want the validator's error")
	}

	empty, err := Open(writeFile(t, "  \n"), fileType)
	if err != nil {
		t.Fatal(err)
	}
	if err := empty.Validate(); err != nil {
		t.Errorf("Validate() on empty document = %v, want nil", err)
	}
}
