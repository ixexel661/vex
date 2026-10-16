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
		Choice{Key: "n", Label: "discard", Run: quitNow},
	)
	return nil
}

func saveAndQuit(ctx Context) error {
	return save(ctx, quitNow)
}

func quitNow(ctx Context) error {
	ctx.Quit()
	return nil
}
