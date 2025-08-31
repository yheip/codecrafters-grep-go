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
	case parser.NodeTypeMatch:
		re = singleMatchRegex(node.Value)
	case parser.NodeTypeCaretAnchor:
		re = singleTransitionRegex(StartOfStringTransitioner{})
	case parser.NodeTypeDollorAnchor:
		re = singleTransitionRegex(EndOfStringTransitioner{})
	case parser.NodeTypeBackreference:
		re = singleTransitionRegex(BackreferenceTransitioner{node.GroupName})
	case parser.NodeTypeAlternation:
		return compileAlternation(node, grpNum)
	case parser.NodeTypeGroup:
		return compileGroup(node, grpNum)
	default:
		return nil, fmt.Errorf("unknown expression type: %T", node)
	}

	processQuantifier(re, node.Quantifier)
	return re, nil
}

// singleTransitionRegex creates a regex with a single transition from start to end
func singleTransitionRegex(tr Transitioner) *CompiledRegex {
	start := NewState()
	end := NewState()
	start.AddTransition(end, tr)

	return &CompiledRegex{start, end}
}

// singleMatchRegex creates a regex that matches a single character
func singleMatchRegex(m parser.Matcher) *CompiledRegex {
	return singleTransitionRegex(CharTransitioner{m})
}

// compileAlternation compiles an alternation node into a CompiledRegex
// i.e a|b
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

		start.AddTransition(subRe.initialState, EpsilonTransitioner{})

		subRe.endingState.AddTransition(end, EpsilonTransitioner{})
	}

	if node.Capturing {
		re.initialState.AddStartingGroup(grpName)
		re.endingState.AddEndingGroup(grpName)
	}

	processQuantifier(re, node.Quantifier)

	return re, nil
}

func compileGroup(node *parser.RegexNode, grpNum *int) (*CompiledRegex, error) {
	var grpName string
	var re *CompiledRegex

	if len(node.Children) == 0 {
		return singleTransitionRegex(EpsilonTransitioner{}), nil
	}

	if node.Capturing {
		if node.GroupName != "" {
			grpName = node.GroupName
		} else {
			grpName = fmt.Sprintf("%d", *grpNum)
			*grpNum++
		}
	}

	for _, child := range node.Children {
		current, err := compile(child, grpNum)
		if err != nil {
			return nil, fmt.Errorf("failed to compile child in group: %w", err)
		}

		if re == nil {
			re = current
		} else {
			re.appendRegex(current)
		}
	}

	if node.Capturing && re != nil {
		re.initialState.AddStartingGroup(grpName)
		re.endingState.AddEndingGroup(grpName)
	}

	processQuantifier(re, node.Quantifier)

	return re, nil
}

// processQuantifier modifies the base regex according to the quantifier
func processQuantifier(base *CompiledRegex, q parser.Quantifier) {
	if q.Plus() {
		withPlus(base)
	} else if q.Optional() {
		withOptional(base)
	}
}

// withPlus modifies the base regex to match one or more times
func withPlus(base *CompiledRegex) {
	start := NewState()
	end := NewState()

	start.AddTransition(base.initialState, EpsilonTransitioner{})
	base.endingState.AddTransition(base.initialState, EpsilonTransitioner{})
	base.endingState.AddTransition(end, EpsilonTransitioner{})

	base.initialState = start
	base.endingState = end
}

// withOptional modifies the base regex to match zero or one time
func withOptional(base *CompiledRegex) {
	base.initialState.AddTransition(base.endingState, EpsilonTransitioner{})
}
