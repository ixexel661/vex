package command

import "fmt"

func appCommands() []Command {
	return []Command{
		{Name: "app.quit", Title: "quit", Run: quit},
	}
}

func quit(ctx Context) error {
	doc := ctx.Editor().Document()
	if !doc.Modified() {
		ctx.Quit()
		return nil
	}
	ctx.Ask(fmt.Sprintf("Save changes to %s?", doc.Name()),
		Choice{Key: "y", Label: "save", Run: saveAndQuit},
		Choice{Key: "n", Label: "discard", Run: quitWithoutSaving},
	)
	return nil
}

func saveAndQuit(ctx Context) error {
	if err := saveDocument(ctx); err != nil {
		return err
	}
	ctx.Quit()
	return nil
}

func quitWithoutSaving(ctx Context) error {
	ctx.Quit()
	return nil
}
