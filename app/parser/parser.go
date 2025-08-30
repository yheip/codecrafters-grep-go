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
