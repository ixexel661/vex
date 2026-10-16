package config

import (
	"errors"
	"fmt"
)

type Config struct {
	// Theme is the name of a file in the themes directory, without ".toml".
	Theme string `toml:"theme"`
	// TabWidth is the number of columns a tab character takes on screen.
	TabWidth int    `toml:"tab-width"`
	Format   Format `toml:"format"`
}

// Format is how documents are formatted. While typing, tab, shift+tab and
// backspace in the indentation use Indent too.
type Format struct {
	Indent         Indent `toml:"indent"`
	SelfCloseEmpty bool   `toml:"self-close-empty"`
	KeepBlankLines bool   `toml:"keep-blank-lines"`
	// WrapTagsLongerThan puts each attribute of a longer tag on its own
	// line, 0 never wraps.
	WrapTagsLongerThan int `toml:"wrap-tags-longer-than"`
}

func Default() Config {
	return Config{
		Theme:    "basic",
		TabWidth: 4,
		Format:   Format{Indent: "  "},
	}
}

// validate resets values out of range to their default and says which.
func (c *Config) validate() error {
	def := Default()
	var errs []error
	if c.Theme == "" {
		errs = append(errs, errors.New("theme must not be empty"))
		c.Theme = def.Theme
	}
	if c.TabWidth < 1 || c.TabWidth > 16 {
		errs = append(errs, fmt.Errorf("tab-width must be between 1 and 16, not %d", c.TabWidth))
		c.TabWidth = def.TabWidth
	}
	if c.Format.WrapTagsLongerThan < 0 {
		errs = append(errs, fmt.Errorf("wrap-tags-longer-than must not be negative, not %d", c.Format.WrapTagsLongerThan))
		c.Format.WrapTagsLongerThan = def.Format.WrapTagsLongerThan
	}
	return errors.Join(errs...)
}
