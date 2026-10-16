package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ixexel661/vex/internal/document"
	"github.com/ixexel661/vex/internal/filetype"
)

var errUsage = errors.New("usage")

// openDocument opens the file the arguments name. With "new" first, it
// creates the file, or without a name a document that asks for one when it
// is saved. isNew says if the document is a new one.
func openDocument(args []string, fileTypes *filetype.Registry) (doc *document.Document, isNew bool, err error) {
	switch {
	case len(args) == 1 && args[0] != "new":
		doc, err := document.Open(args[0], fileTypes.Detect(args[0]))
		return doc, false, err
	case len(args) == 1:
		return document.New("", fileTypes.Default()), true, nil
	case len(args) == 2 && args[0] == "new":
		doc, err := createDocument(args[1], fileTypes, os.Stdin, os.Stderr)
		return doc, true, err
	default:
		return nil, false, errUsage
	}
}

// createDocument writes a new file with the template of its type. A name
// without an extension gets the one of the default type. If there is a file
// already, it asks before overwriting it.
func createDocument(name string, fileTypes *filetype.Registry, in io.Reader, out io.Writer) (*document.Document, error) {
	path := fileTypes.Default().FileName(name)
	if document.Exists(path) && !confirm(in, out, path+" exists already. Overwrite it?") {
		return nil, fmt.Errorf("%s exists already", path)
	}
	doc := document.New(path, fileTypes.Detect(path))
	if err := doc.Save(); err != nil {
		return nil, err
	}
	return doc, nil
}

// confirm asks a yes or no question. Anything but y or yes is no.
func confirm(in io.Reader, out io.Writer, question string) bool {
	fmt.Fprintf(out, "%s [y/N] ", question)
	answer, _ := bufio.NewReader(in).ReadString('\n')
	answer = strings.ToLower(strings.TrimSpace(answer))
	return answer == "y" || answer == "yes"
}
