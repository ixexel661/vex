package command

func settingsCommands() []Command {
	return []Command{
		{Name: "settings.open", Title: "settings", Run: openSettings},
	}
}

func openSettings(ctx Context) error {
	ctx.OpenSettings()
	return nil
}
