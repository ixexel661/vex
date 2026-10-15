package command

import "github.com/ixexel661/vex/internal/mode"

func modeCommands() []Command {
	return []Command{
		{Name: "mode.insert", Title: "insert", Run: switchTo(mode.Insert)},
		{Name: "mode.normal", Title: "normal mode", Run: switchTo(mode.Normal)},
	}
}

func switchTo(m mode.Mode) func(Context) error {
	return func(ctx Context) error {
		ctx.SetMode(m)
		return nil
	}
}
