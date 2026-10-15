package command

import "github.com/ixexel661/vex/internal/editor"

func editCommands() []Command {
	return []Command{
		{Name: "edit.newline", Run: onEditor((*editor.Editor).Newline)},
		{Name: "edit.delete-backward", Run: onEditor((*editor.Editor).DeleteBackward)},
		{Name: "edit.delete-forward", Run: onEditor((*editor.Editor).DeleteForward)},
		{Name: "edit.indent", Run: indent},
	}
}

func indent(ctx Context) error {
	ctx.Editor().InsertText(ctx.Config().Indent)
	return nil
}
