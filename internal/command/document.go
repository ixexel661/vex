package command

import "fmt"

func documentCommands() []Command {
	return []Command{
		{Name: "document.save", Title: "save", Run: saveDocument},
		{Name: "document.format", Title: "format", Run: formatDocument},
	}
}

func saveDocument(ctx Context) error {
	doc := ctx.Editor().Document()
	if err := doc.Save(); err != nil {
		return err
	}
	ctx.Notify(fmt.Sprintf("Saved %s", doc.Name()))
	return nil
}

func formatDocument(ctx Context) error {
	doc := ctx.Editor().Document()
	if doc.Type.Formatter == nil {
		return fmt.Errorf("%s files cannot be formatted", doc.Type.Name)
	}
	text := doc.Buffer.Text()
	formatted, err := doc.Type.Formatter.Format(text, ctx.Config().Format)
	if err != nil {
		return fmt.Errorf("cannot format: %w", err)
	}
	if formatted == text {
		ctx.Notify("Already formatted")
		return nil
	}
	ctx.Editor().SetText(formatted)
	ctx.Notify("Formatted")
	return nil
}
