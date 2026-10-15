package config

type Config struct {
	// TabWidth is the number of columns a tab character takes on screen.
	TabWidth int
	// Indent is inserted by the tab key and used to format documents.
	Indent string
}

func Default() Config {
	return Config{TabWidth: 4, Indent: "  "}
}
