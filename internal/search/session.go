package search

import "github.com/ixexel661/vex/internal/buffer"

// Session is a search the user refines while it runs. It is kept between
// searches, so the next one starts with the same query and matcher.
type Session struct {
	lines    Lines
	matchers []Matcher
	matcher  int // index of the current matcher

	query    string
	origin   buffer.Position
	matches  []Match
	selected int
}

// NewSession searches lines with the given matchers. The first one is used
// until NextMatcher switches to another one.
func NewSession(lines Lines, matchers []Matcher) *Session {
	return &Session{lines: lines, matchers: matchers, selected: -1}
}

// Start begins a search at from with the query of the last one.
func (s *Session) Start(from buffer.Position) {
	s.origin = from
	s.find()
}

// Origin returns where the search started.
func (s *Session) Origin() buffer.Position {
	return s.origin
}

func (s *Session) Query() string {
	return s.query
}

func (s *Session) SetQuery(query string) {
	s.query = query
	s.find()
}

func (s *Session) Matcher() Matcher {
	return s.matchers[s.matcher]
}

// NextMatcher searches again with the next matcher, after the last one
// comes the first one.
func (s *Session) NextMatcher() {
	s.matcher = (s.matcher + 1) % len(s.matchers)
	s.find()
}

func (s *Session) Matches() []Match {
	return s.matches
}

// Selected returns the index of the selected match, or -1 if there are no
// matches.
func (s *Session) Selected() int {
	return s.selected
}

// Move selects the match delta steps away and wraps around at both ends.
func (s *Session) Move(delta int) {
	if n := len(s.matches); n > 0 {
		s.selected = ((s.selected+delta)%n + n) % n
	}
}

func (s *Session) find() {
	s.matches = s.Matcher().Find(s.lines, s.query)
	s.selected = s.first()
}

// first returns the match to select after a search: the best one if the
// matcher ranks its matches, otherwise the first one at or after the origin.
// If there is none after it, the search wraps around to the top.
func (s *Session) first() int {
	if len(s.matches) == 0 {
		return -1
	}
	if s.Matcher().Ranked() {
		return 0
	}
	for i, m := range s.matches {
		if !m.Position().Before(s.origin) {
			return i
		}
	}
	return 0
}
