package xml

import "github.com/ixexel661/vex/internal/filetype"

func Type() filetype.Type {
	return filetype.Type{
		Name: "XML",
		Extensions: []string{
			".xml", ".xsd", ".xsl", ".xslt", ".svg", ".xhtml",
			".rss", ".atom", ".wsdl", ".plist", ".csproj", ".xaml",
		},
		Highlighter: Highlighter{},
		Formatter:   Formatter{},
		Validator:   Validator{},
	}
}
