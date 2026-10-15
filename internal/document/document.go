package document

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/ixexel661/vex/internal/buffer"
	"github.com/ixexel661/vex/internal/filetype"
)

const bom = "\uFEFF"

// Document connects a file on disk with the buffer it is edited in.
type Document struct {
	Path   string
	Type   filetype.Type
	Buffer *buffer.Buffer

	lineEnding   string
	hasBOM       bool
	savedVersion int
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
	lineEnding := "\n"
	if strings.Contains(text, "\r\n") {
		lineEnding = "\r\n"
		text = strings.ReplaceAll(text, "\r\n", "\n")
	}
	return &Document{
		Path:       path,
		Type:       fileType,
		Buffer:     buffer.New(text),
		lineEnding: lineEnding,
		hasBOM:     hasBOM,
	}, nil
}

func (d *Document) Name() string {
	return filepath.Base(d.Path)
}

func (d *Document) Modified() bool {
	return d.Buffer.Version() != d.savedVersion
}

func (d *Document) Save() error {
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

// Validate checks the text with the validator of the file type. Empty
// documents and file types without a validator count as valid.
func (d *Document) Validate() error {
	text := d.Buffer.Text()
	if d.Type.Validator == nil || strings.TrimSpace(text) == "" {
		return nil
	}
	return d.Type.Validator.Validate(text)
}
