package xml

import "github.com/ixexel661/vex/internal/filetype"

func Type() filetype.Type {
	return filetype.Type{
		Name: "XML",
		Extensions: []string{
			".xml", ".xsd", ".xsl", ".xslt", ".svg", ".xhtml",
			".rss", ".atom", ".wsdl", ".plist", ".csproj", ".xaml",
		},
		Template:    "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n",
		Highlighter: Highlighter{},
		Formatter:   Formatter{},
		Validator:   Validator{},
		Tags:        Tags{},
	}
}
