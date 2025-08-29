package parser

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
	NodeTypeLiteral NodeType = iota
	NodeTypeAlternation
	NodeTypeGroup
)

type RegexNode struct {
	Type       NodeType
	Children   []*RegexNode // For alternatives, groups
	Value      byte         // For literals
	Quantifier Quantifier
	Capturing  bool
	GroupName  string
}

func (n *RegexNode) WithQuantifier(q Quantifier) *RegexNode {
	n.Quantifier = q

	return n
}

func NewLiteral(value byte) *RegexNode {
	return &RegexNode{
		Type:  NodeTypeLiteral,
		Value: value,
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
