package document

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
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

func TestNew(t *testing.T) {
	doc := New("", filetype.Type{Template: "<?xml?>\n"})
	if got := doc.Buffer.Text(); got != "<?xml?>\n" {
		t.Errorf("text = %q, want the template", got)
	}
	if doc.Modified() {
		t.Error("new document is modified")
	}
	if got := doc.Name(); got != "untitled" {
		t.Errorf("Name() = %q, want untitled", got)
	}
	if err := doc.Save(); err == nil {
		t.Error("Save() without a path succeeded")
	}
}

func TestNameHasNoControlCharacters(t *testing.T) {
	doc := New("/tmp/a\a\x1b]52;c;eA==\a.xml", filetype.Type{})
	if got, want := doc.Name(), "a??]52;c;eA==?.xml"; got != want {
		t.Errorf("Name() = %q, want %q", got, want)
	}
}

func TestSaveAs(t *testing.T) {
	dir := t.TempDir()
	doc := New("", filetype.Plain())
	doc.Buffer.Insert(buffer.Position{}, "<a/>")

	if err := doc.SaveAs(filepath.Join(dir, "missing", "a.xml")); err == nil {
		t.Fatal("SaveAs() into a missing directory succeeded")
	}
	if doc.Path != "" || !doc.Modified() {
		t.Errorf("after a failed SaveAs() path = %q, modified = %v, want them unchanged", doc.Path, doc.Modified())
	}

	path := filepath.Join(dir, "a.xml")
	if err := doc.SaveAs(path); err != nil {
		t.Fatal(err)
	}
	if got := readFile(t, path); got != "<a/>" {
		t.Errorf("saved file = %q", got)
	}
	if doc.Name() != "a.xml" || doc.Modified() {
		t.Errorf("name = %q, modified = %v, want a.xml and saved", doc.Name(), doc.Modified())
	}
}

func TestExists(t *testing.T) {
	path := writeFile(t, "")
	if !Exists(path) {
		t.Errorf("Exists(%q) = false", path)
	}
	if missing := filepath.Join(filepath.Dir(path), "missing.xml"); Exists(missing) {
		t.Errorf("Exists(%q) = true", missing)
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

func TestValidationChecksTheTextFromItsStart(t *testing.T) {
	doc := New("", filetype.Type{Name: "Strict", Validator: rejectAll{}})
	check := doc.Validation()
	doc.Buffer.SetText("<a/>")
	if err := check().Err; err != nil {
		t.Errorf("the check of the empty text saw the text after it: %v", err)
	}
}

// copiesLines reports whether an edit copies the list of lines, which the
// buffer does while a snapshot is not released.
func copiesLines(doc *Document) bool {
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	doc.Buffer.Insert(buffer.Position{}, "x")
	runtime.ReadMemStats(&after)
	return after.TotalAlloc-before.TotalAlloc > uint64(doc.Buffer.LineCount()*8)
}

func TestValidationReleasesItsSnapshot(t *testing.T) {
	doc := New("", filetype.Type{Name: "Strict", Validator: rejectAll{}})
	doc.Buffer.SetText(strings.Repeat("<a/>\n", 10_000))
	doc.Validation()
	if !copiesLines(doc) {
		t.Fatal("an edit does not copy the lines while the check reads them")
	}
	doc.Validation()()
	if copiesLines(doc) {
		t.Error("an edit copies the lines after the check is done")
	}
}
