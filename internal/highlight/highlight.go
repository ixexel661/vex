package highlight

// Kind classifies a piece of text, so a theme can pick its color.
type Kind int

const (
	Plain Kind = iota
	Punctuation
	Tag
	Attribute
	String
	Comment
	Meta
	Entity
	CData
	Match
	CurrentMatch
	LinkedTag
)

// Span marks the runes from Start up to, but not including, End of a line.
type Span struct {
	Start, End int
	Kind       Kind
}

// State is what a highlighter carries from the end of one line to the start
// of the next one, e.g. "inside a comment". Every highlighter defines its own
// states; zero is the state at the start of a document.
type State int

type Highlighter interface {
	Highlight(line []rune, state State) ([]Span, State)
}
