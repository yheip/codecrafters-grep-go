package regex

import (
	"fmt"

	"github.com/codecrafters-io/grep-starter-go/app/parser"
)

func Compile(root *parser.RegexNode) (*CompiledRegex, error) {
	return compile(root, new(int))
}

func compile(node *parser.RegexNode, grpNum *int) (*CompiledRegex, error) {
	var re *CompiledRegex
	switch node.Type {
	case parser.NodeTypeLiteral:
		re = NewSingleCharRegex(node.Value)
		processQuantifier(re, node.Quantifier)

		return re, nil
	case parser.NodeTypeGroup:
		var grpName string

		if node.Capturing {
			if node.GroupName != "" {
				grpName = node.GroupName
			} else {
				grpName = fmt.Sprintf("%d", *grpNum)
				*grpNum++
			}
		}

		for _, child := range node.Children {
			var current *CompiledRegex
			switch child.Type {
			case parser.NodeTypeLiteral:
				current = NewSingleCharRegex(child.Value)
			case parser.NodeTypeAlternation:
				var err error
				current, err = compileAlternation(child, grpNum)
				if err != nil {
					return nil, fmt.Errorf("failed to compile alteration: %w", err)
				}
			case parser.NodeTypeGroup:
				var err error
				current, err = compile(child, grpNum)
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

		if node.Capturing {
			re.initialState.AddStartingGroup(grpName)
			re.endingState.AddEndingGroup(grpName)
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
func compileAlternation(node *parser.RegexNode, grpNum *int) (*CompiledRegex, error) {
	var grpName string
	if node.Capturing {
		if node.GroupName != "" {
			grpName = node.GroupName
		} else {
			grpName = fmt.Sprintf("%d", *grpNum)
			*grpNum++
		}
	}

	start, end := NewState(), NewState()
	re := &CompiledRegex{start, end}

	// Add an union for each alternative
	for _, child := range node.Children {
		var subRe *CompiledRegex
		var err error

		subRe, err = compile(child, grpNum)
		if err != nil {
			return nil, err
		}

		start.AddTransition(subRe.initialState, EpsilonMatcher{})

		subRe.endingState.AddTransition(end, EpsilonMatcher{})
	}

	if node.Capturing {
		re.initialState.AddStartingGroup(grpName)
		re.endingState.AddEndingGroup(grpName)
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
