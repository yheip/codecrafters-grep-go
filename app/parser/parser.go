package parser

import "slices"

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
	// TODO
	return nil, nil
}

type NodeType int

const (
	NodeTypeMatch NodeType = iota
	NodeTypeAlternation
	NodeTypeGroup
)

type RegexNode struct {
	Type       NodeType
	Children   []*RegexNode // For alternations, groups
	Value      Matcher      // For match nodes
	Quantifier Quantifier
	Capturing  bool
	GroupName  string
}

func (n *RegexNode) WithQuantifier(q Quantifier) *RegexNode {
	n.Quantifier = q

	return n
}

func NewLiteralMatch(value byte) *RegexNode {
	return &RegexNode{
		Type:  NodeTypeMatch,
		Value: &LiteralMatcher{Char: value},
	}
}

func NewCharGroupMatch(m *CharGroupMatcher) *RegexNode {
	return &RegexNode{
		Type:  NodeTypeMatch,
		Value: m,
	}
}

func NewGroup(children []*RegexNode) *RegexNode {
	return &RegexNode{
		Type:     NodeTypeGroup,
		Children: children,
	}
}

func NewAlternation(alternatives []*RegexNode) *RegexNode {
	return &RegexNode{
		Type:     NodeTypeAlternation,
		Children: alternatives,
	}
}

type Quantifier int

const (
	QuantifierNone Quantifier = 1 << iota
	QuantifierAsterisk
	QuantifierPlus
	QuantifierOptional
)

func (q Quantifier) None() bool {
	return q&QuantifierNone != 0
}

func (q Quantifier) Plus() bool {
	return q&QuantifierPlus != 0
}

func (q Quantifier) Optional() bool {
	return q&QuantifierOptional != 0
}

type Matcher interface {
	Match(c byte) bool
	String() string
}

type LiteralMatcher struct {
	Char byte
}

func (m *LiteralMatcher) Match(c byte) bool {
	return m.Char == c
}

func (m *LiteralMatcher) String() string {
	return string(m.Char)
}

type CharGroupMatcher struct {
	Chars  []byte
	Ranges [][2]byte
	Negate bool
	Label  string
}

func (m *CharGroupMatcher) Match(c byte) bool {
	found := slices.Contains(m.Chars, c)

	if !found {
		found = slices.ContainsFunc(m.Ranges, func(r [2]byte) bool {
			return c >= r[0] && c <= r[1]
		})
	}

	if m.Negate {
		return !found
	}

	return found
}

func (m *CharGroupMatcher) String() string {
	return m.Label
}

var (
	DigitMatcher = &CharGroupMatcher{
		Chars: []byte{},
		Ranges: [][2]byte{
			{'0', '9'},
		},
		Label: `\d`,
	}
	WordMatcher = &CharGroupMatcher{
		Chars: []byte{'_'},
		Ranges: [][2]byte{
			{'a', 'z'},
			{'A', 'Z'},
			{'0', '9'},
		},
		Label: `\w`,
	}
)
