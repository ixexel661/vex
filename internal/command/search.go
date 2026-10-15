package command

func searchCommands() []Command {
	return []Command{
		{Name: "search.open", Title: "search", Run: openSearch},
	}
}

func openSearch(ctx Context) error {
	ctx.OpenSearch()
	return nil
}
