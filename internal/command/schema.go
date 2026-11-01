package command

func schemaCommands() []Command {
	return []Command{
		{Name: "schema.problems", Title: "errors", Run: showProblems},
		{Name: "schema.edit", Title: "schema", Run: openSchemas},
	}
}

func showProblems(ctx Context) error {
	ctx.ShowProblems()
	return nil
}

func openSchemas(ctx Context) error {
	ctx.OpenSchemas()
	return nil
}
