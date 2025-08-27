package regex

import (
	"fmt"

	"github.com/codecrafters-io/grep-starter-go/app/parser"
)

func Compile(root *parser.RegexNode) (*CompiledRegex, error) {
	var re *CompiledRegex
	switch root.Type {
	case parser.NodeTypeLiteral:
		re = NewSingleCharRegex(root.Value)
		processQuantifier(re, root.Quantifier)

		return re, nil
	case parser.NodeTypeGroup:
		for _, child := range root.Children {
			var current *CompiledRegex
			switch child.Type {
			case parser.NodeTypeLiteral:
				current = NewSingleCharRegex(child.Value)
			case parser.NodeTypeAlternation:
				var err error
				current, err = compileAlternation(child)
				if err != nil {
					return nil, fmt.Errorf("failed to compile alteration: %w", err)
				}
			case parser.NodeTypeGroup:
				var err error
				current, err = Compile(child)
				if err != nil {
					return nil, fmt.Errorf("failed to compile child group: %w", err)
				}
			default:
				return nil, fmt.Errorf("unknown expression type in group: %T", child)
			}

			processQuantifier(current, child.Quantifier)

			if re == nil {
				re = current
			} else {
				// Concatenate re and current
				re.appendRegex(current)
			}
		}
	}

	return re, nil
}

func NewSingleCharRegex(c byte) *CompiledRegex {
	start := NewState()
	end := NewState()
	start.AddTransition(end, CharMatcher{Char: c})

	return &CompiledRegex{start, end}
}

// a|b
func compileAlternation(node *parser.RegexNode) (*CompiledRegex, error) {
	start, end := NewState(), NewState()
	re := &CompiledRegex{start, end}

	// Add an union for each alternative
	for _, child := range node.Children {
		var subRe *CompiledRegex
		var err error

		subRe, err = Compile(child)
		if err != nil {
			return nil, err
		}

		for _, tr := range subRe.initialState.Transitions {
			start.AddTransition(tr.To, tr.Matcher)
		}

		subRe.endingState.AddTransition(end, EpsilonMatcher{})
	}

	return re, nil
}

func processQuantifier(base *CompiledRegex, q parser.Quantifier) {
	if q.Plus() {
		withPlus(base)
	} else if q.Optional() {
		withOptional(base)
	}
}

func withPlus(base *CompiledRegex) {
	start := NewState()
	end := NewState()

	start.AddTransition(base.initialState, EpsilonMatcher{})
	base.endingState.AddTransition(base.initialState, EpsilonMatcher{})
	base.endingState.AddTransition(end, EpsilonMatcher{})

	base.initialState = start
	base.endingState = end
}

func withOptional(base *CompiledRegex) {
	base.initialState.AddTransition(base.endingState, EpsilonMatcher{})
}
