package document

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/ixexel661/vex/internal/buffer"
	"github.com/ixexel661/vex/internal/filetype"
	"github.com/ixexel661/vex/internal/printable"
)

const bom = "\uFEFF"

// Document connects a file on disk with the buffer it is edited in.
type Document struct {
	Path   string // empty until a new document is saved for the first time
	Type   filetype.Type
	Buffer *buffer.Buffer

	lineEnding   string
	hasBOM       bool
	savedVersion int
	schemas      *schemaSet
}

// Open reads the file at path. A missing file gives an empty document,
// the file is created on the first save.
func Open(path string, fileType filetype.Type) (*Document, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		data, err = nil, nil
	}
	if err != nil {
		return nil, err
	}
	if !utf8.Valid(data) {
		return nil, fmt.Errorf("%s is not UTF-8 encoded, which is the only encoding vex supports", path)
	}

	text, hasBOM := strings.CutPrefix(string(data), bom)
	newBuffer, lineEnding := buffer.New, "\n"
	if strings.Contains(text, "\r\n") {
		newBuffer, lineEnding = buffer.NewCRLF, "\r\n"
	}
	return &Document{
		Path:       path,
		Type:       fileType,
		Buffer:     newBuffer(text),
		lineEnding: lineEnding,
		hasBOM:     hasBOM,
	}, nil
}

// New returns a document for a new file at path, with the template of its
// type. Nothing is written yet, and path may be empty.
func New(path string, fileType filetype.Type) *Document {
	return &Document{
		Path:       path,
		Type:       fileType,
		Buffer:     buffer.New(fileType.Template),
		lineEnding: "\n",
	}
}

// Exists reports whether there is a file at path.
func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// Name is the file name to show. Control characters in it are replaced, so
// a file name cannot send escape sequences to the terminal.
func (d *Document) Name() string {
	if d.Path == "" {
		return "untitled"
	}
	return printable.String(filepath.Base(d.Path))
}

func (d *Document) Modified() bool {
	return d.Buffer.Version() != d.savedVersion
}

// MarkUnmodified says the text is the one in the file again, e.g. after a
// change was taken back.
func (d *Document) MarkUnmodified() {
	d.savedVersion = d.Buffer.Version()
}

func (d *Document) Save() error {
	if d.Path == "" {
		return errors.New("the document has no file name")
	}
	text := strings.ReplaceAll(d.Buffer.Text(), "\n", d.lineEnding)
	if d.hasBOM {
		text = bom + text
	}
	if err := os.WriteFile(d.Path, []byte(text), 0o644); err != nil {
		return err
	}
	d.savedVersion = d.Buffer.Version()
	return nil
}

// SaveAs saves the document at path, which it keeps from then on. If saving
// fails, the document keeps its old path.
func (d *Document) SaveAs(path string) error {
	old := d.Path
	d.Path = path
	if err := d.Save(); err != nil {
		d.Path = old
		return err
	}
	return nil
}

// Validity is the result of a check of a document.
type Validity struct {
	Err error // why the text is not well-formed
	// Schema reports whether the text was checked against schemas.
	// SchemaErr says why they could not be loaded, Problems what is wrong
	// in the text against them.
	Schema    bool
	SchemaErr error
	Problems  []filetype.Problem
}

// Validate checks that the text is well-formed with the validator of the
// file type. Empty documents and file types without a validator count as
// valid.
func (d *Document) Validate() error {
	return d.Validation()().Err
}

// Validation returns a check of the text as it is now, which can run in
// another goroutine while the text changes. It is meant to run once. With
// schemas, it checks the text against them at the same time.
func (d *Document) Validation() func() Validity {
	validator, snapshot := d.Type.Validator, d.Buffer.Snapshot()
	schemaValidator, schemas := d.Type.Schemas, d.schemas
	return func() Validity {
		text := snapshot.Text()
		snapshot.Release()
		v := Validity{Schema: schemas != nil}
		if strings.TrimSpace(text) == "" {
			return v
		}
		var wg sync.WaitGroup
		if schemas != nil {
			wg.Go(func() {
				check, err := schemas.load(schemaValidator)
				if err != nil {
					v.SchemaErr = err
					return
				}
				v.Problems = check(text)
			})
		}
		if validator != nil {
			v.Err = validator.Validate(text)
		}
		wg.Wait()
		if v.Err != nil {
			v.SchemaErr, v.Problems = nil, nil
		}
		return v
	}
}
