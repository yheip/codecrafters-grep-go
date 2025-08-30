package parser

import "fmt"

type Parser struct {
	pattern string
	pos     int
}

func New(pattern string) *Parser {
	return &Parser{
		pattern: pattern,
		pos:     0,
	}
}

func (p *Parser) Parse() (*RegexNode, error) {
	// Parse the whole pattern and wrap it in a top-level capturing group
	alt, seq, err := p.parseAlternation('\x00') // no explicit stop char
	if err != nil {
		return nil, err
	}

	// Ensure entire pattern was consumed
	if p.pos < len(p.pattern) {
		return nil, p.errorf("unexpected character '%c' at position %d", p.pattern[p.pos], p.pos)
	}

	// Top-level must be a capturing group
	if alt != nil {
		// Place the alternation as a single child of the top-level group
		return &RegexNode{Type: NodeTypeGroup, Children: []*RegexNode{alt}, Capturing: true}, nil
	}

	return &RegexNode{Type: NodeTypeGroup, Children: seq, Capturing: true}, nil
}

// parseAlternation parses sequences separated by '|'.
// If an alternation is present, it returns a NodeTypeAlternation node (alt != nil) and seq=nil.
// If there is no alternation, it returns alt=nil and the parsed sequence as seq.
func (p *Parser) parseAlternation(stop byte) (alt *RegexNode, seq []*RegexNode, err error) {
	var alternatives [][]*RegexNode

	for {
		s, e := p.parseSequence(stop)
		if e != nil {
			return nil, nil, e
		}
		alternatives = append(alternatives, s)

		// If current char is '|', consume and parse next alternative
		if p.peek() == '|' {
			p.next()
			continue
		}
		break
	}

	if len(alternatives) == 1 {
		return nil, alternatives[0], nil
	}

	// Build an alternation node. Each alternative sequence becomes a child. If a
	// sequence has multiple terms, wrap them in a non-capturing group to preserve order.
	children := make([]*RegexNode, 0, len(alternatives))
	for _, s := range alternatives {
		if len(s) == 1 {
			children = append(children, s[0])
		} else {
			children = append(children, NewGroup(s))
		}
	}

	return NewAlternation(children), nil, nil
}

// parseSequence parses a sequence of terms until stop char, '|' or ')' (if stop is ')').
func (p *Parser) parseSequence(stop byte) ([]*RegexNode, error) {
	nodes := []*RegexNode{}

	for {
		c := p.peek()
		if c == 0 || c == '|' || (stop != 0 && c == stop) {
			break
		}

		node, err := p.parseTerm(stop)
		if err != nil {
			return nil, err
		}
		if node != nil {
			nodes = append(nodes, node)
		} else {
			break
		}
	}

	return nodes, nil
}

// parseTerm parses a single term and its quantifier if present.
func (p *Parser) parseTerm(stop byte) (*RegexNode, error) {
	c := p.peek()
	if c == 0 || c == '|' || (stop != 0 && c == stop) {
		return nil, nil
	}

	var node *RegexNode

	switch c {
	case '^':
		p.next()
		node = NewCaretAnchor()
	case '$':
		p.next()
		node = NewDollarAnchor()
	case '.':
		p.next()
		node = NewCharGroupMatch(WildcardMatcher)
	case '(':
		n, err := p.parseGroup()
		if err != nil {
			return nil, err
		}
		node = n
	case '[':
		n, err := p.parseCharClass()
		if err != nil {
			return nil, err
		}
		node = n
	case '\\':
		p.next()
		if p.eof() {
			return nil, p.errorf("incomplete escape at end of pattern")
		}
		esc := p.next()
		switch esc {
		case 'd':
			node = NewCharGroupMatch(DigitMatcher)
		case 'w':
			node = NewCharGroupMatch(WordMatcher)
		default:
			node = NewLiteralMatch(esc)
		}
	default:
		// literal character
		p.next()
		node = NewLiteralMatch(c)
	}

	// Parse optional quantifier
	switch p.peek() {
	case '+':
		node.Quantifier |= QuantifierPlus
		p.next()
	case '?':
		node.Quantifier |= QuantifierOptional
		p.next()
	}

	return node, nil
}

// parseGroup parses a capturing group: '(' ... ')'
func (p *Parser) parseGroup() (*RegexNode, error) {
	// consume '('
	if p.next() != '(' {
		return nil, p.errorf("expected '(' at position %d", p.pos-1)
	}

	alt, seq, err := p.parseAlternation(')')
	if err != nil {
		return nil, err
	}

	if p.peek() != ')' {
		return nil, p.errorf("unmatched '(' at position %d", p.pos-1)
	}
	// consume ')'
	p.next()

	var node *RegexNode
	if alt != nil {
		alt.Capturing = true
		node = alt
	} else {
		node = NewGroup(seq)
		node.Capturing = true
	}

	// optional quantifier after group
	switch p.peek() {
	case '+':
		node.Quantifier |= QuantifierPlus
		p.next()
	case '?':
		node.Quantifier |= QuantifierOptional
		p.next()
	}

	return node, nil
}

// parseCharClass parses a character class like [abc] or [^abc]. No range parsing required.
func (p *Parser) parseCharClass() (*RegexNode, error) {
	if p.next() != '[' { // consume '['
		return nil, p.errorf("expected '[' at position %d", p.pos-1)
	}

	negate := false
	if p.peek() == '^' {
		negate = true
		p.next()
	}

	cg := &CharGroupMatcher{Chars: []byte{}, Ranges: [][2]byte{}, Negate: negate}

	// collect until ']'
	for {
		if p.eof() {
			return nil, p.errorf("unmatched '[' in character class")
		}
		ch := p.next()
		if ch == ']' {
			break
		}
		if ch == '\\' {
			if p.eof() {
				return nil, p.errorf("incomplete escape in character class")
			}
			// Only treat \] or \\ specially; \d/\w inside class are not supported here, treat as literals
			ch = p.next()
			cg.Chars = append(cg.Chars, ch)
			continue
		}

		// Minimal range support: a-b
		if p.peek() == '-' {
			// lookahead next-next to ensure it's a proper range like a-b]
			// consume '-'
			p.next()
			if p.eof() {
				return nil, p.errorf("unterminated range in character class")
			}
			end := p.next()
			if end == ']' {
				// Treat trailing '-' as literal if immediately before closing bracket
				cg.Chars = append(cg.Chars, ch, '-')
				// Put back the ']' by stepping back one pos since we consumed it as end
				p.pos--
				continue
			}
			if end < ch {
				// swap to ensure valid range
				ch, end = end, ch
			}
			cg.Ranges = append(cg.Ranges, [2]byte{ch, end})
			continue
		}

		cg.Chars = append(cg.Chars, ch)
	}

	// optional quantifier after class
	node := NewCharGroupMatch(cg)
	switch p.peek() {
	case '+':
		node.Quantifier |= QuantifierPlus
		p.next()
	case '?':
		node.Quantifier |= QuantifierOptional
		p.next()
	}

	return node, nil
}

// helpers
func (p *Parser) peek() byte {
	if p.pos >= len(p.pattern) {
		return 0
	}
	return p.pattern[p.pos]
}

func (p *Parser) next() byte {
	if p.pos >= len(p.pattern) {
		return 0
	}
	ch := p.pattern[p.pos]
	p.pos++
	return ch
}

func (p *Parser) eof() bool { return p.pos >= len(p.pattern) }

func (p *Parser) errorf(format string, args ...any) error {
	return fmt.Errorf(format, args...)
}
