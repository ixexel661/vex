package command

import (
	"fmt"
	"path/filepath"

	"github.com/ixexel661/vex/internal/document"
)

func documentCommands() []Command {
	return []Command{
		{Name: "document.save", Title: "save", Run: saveDocument},
		{Name: "document.format", Title: "format", Run: formatDocument},
	}
}

func saveDocument(ctx Context) error {
	return save(ctx, done)
}

// save saves the document and runs next after it. A document without a file
// asks for its name first, and if there is a file of that name already,
// whether to overwrite it.
func save(ctx Context, next func(Context) error) error {
	doc := ctx.Editor().Document()
	if doc.Path != "" {
		return saveAt(doc.Path, next)(ctx)
	}
	ctx.Prompt("Save as:", func(ctx Context, name string) error {
		path := doc.Type.FileName(name)
		if !document.Exists(path) {
			return saveAt(path, next)(ctx)
		}
		ctx.Ask(fmt.Sprintf("%s exists already. Overwrite it?", filepath.Base(path)),
			Choice{Key: "y", Label: "overwrite", Run: saveAt(path, next)},
		)
		return nil
	})
	return nil
}

// saveAt saves the document at path and runs next after it.
func saveAt(path string, next func(Context) error) func(Context) error {
	return func(ctx Context) error {
		doc := ctx.Editor().Document()
		if err := doc.SaveAs(path); err != nil {
			return err
		}
		ctx.Notify(fmt.Sprintf("Saved %s", doc.Name()))
		return next(ctx)
	}
}

func done(Context) error {
	return nil
}

func formatDocument(ctx Context) error {
	doc := ctx.Editor().Document()
	text := doc.Buffer.Text()
	formatted, err := doc.Type.Format(text, ctx.Config().Format)
	if err != nil {
		return err
	}
	if formatted == text {
		ctx.Notify("Already formatted")
		return nil
	}
	ctx.Editor().SetText(formatted)
	ctx.Notify("Formatted")
	return nil
}
