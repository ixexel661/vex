package command

import (
	"slices"

	"github.com/ixexel661/vex/internal/editor"
)

// Builtins returns all commands that ship with vex.
func Builtins() []Command {
	return slices.Concat(
		appCommands(),
		modeCommands(),
		documentCommands(),
		searchCommands(),
		settingsCommands(),
		cursorCommands(),
		editCommands(),
	)
}

// onEditor turns an editor method into a command.
func onEditor(action func(*editor.Editor)) func(Context) error {
	return func(ctx Context) error {
		action(ctx.Editor())
		return nil
	}
}
