package parser

import (
	"testing"
)

// nodesEqual recursively compares two RegexNode trees for semantic equality.
func nodesEqual(a, b *RegexNode) bool {
	if a == nil || b == nil {
		return a == b
	}

	if a.Type != b.Type || a.Quantifier != b.Quantifier || a.Capturing != b.Capturing || a.GroupName != b.GroupName {
		return false
	}

	// Compare Values for match nodes
	switch av := a.Value.(type) {
	case *LiteralMatcher:
		bv, ok := b.Value.(*LiteralMatcher)
		if !ok || av.Char != bv.Char {
			return false
		}
	case *CharGroupMatcher:
		bv, ok := b.Value.(*CharGroupMatcher)
		if !ok {
			return false
		}
		// If labels are set (e.g., \d, \w), compare labels first
		if av.Label != "" || bv.Label != "" {
			if av.Label != bv.Label {
				return false
			}
		} else {
			if av.Negate != bv.Negate {
				return false
			}
			if len(av.Chars) != len(bv.Chars) {
				return false
			}
			for i := range av.Chars {
				if av.Chars[i] != bv.Chars[i] {
					return false
				}
			}
			if len(av.Ranges) != len(bv.Ranges) {
				return false
			}
			for i := range av.Ranges {
				if av.Ranges[i] != bv.Ranges[i] {
					return false
				}
			}
		}
	default:
		if a.Value != nil || b.Value != nil {
			return false
		}
	}

	if len(a.Children) != len(b.Children) {
		return false
	}
	for i := range a.Children {
		if !nodesEqual(a.Children[i], b.Children[i]) {
			return false
		}
	}

	return true
}

func TestParser_Parse_Basic(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		want    func() *RegexNode
	}{
		{
			name:    "backreference \\1",
			pattern: "\\1",
			want: func() *RegexNode {
				return &RegexNode{Type: NodeTypeGroup, Capturing: true, Children: []*RegexNode{
					NewBackreference("1"),
				}}
			},
		},
		{
			name:    "backreference multi-digit \\12",
			pattern: "\\12",
			want: func() *RegexNode {
				return &RegexNode{Type: NodeTypeGroup, Capturing: true, Children: []*RegexNode{
					NewBackreference("12"),
				}}
			},
		},
		{
			name:    "single literal",
			pattern: "a",
			want: func() *RegexNode {
				return &RegexNode{Type: NodeTypeGroup, Capturing: true, Children: []*RegexNode{
					NewLiteralMatch('a'),
				}}
			},
		},
		{
			name:    "digits \\d",
			pattern: "\\d",
			want: func() *RegexNode {
				return &RegexNode{Type: NodeTypeGroup, Capturing: true, Children: []*RegexNode{
					NewCharGroupMatch(DigitMatcher),
				}}
			},
		},
		{
			name:    "word \\w",
			pattern: "\\w",
			want: func() *RegexNode {
				return &RegexNode{Type: NodeTypeGroup, Capturing: true, Children: []*RegexNode{
					NewCharGroupMatch(WordMatcher),
				}}
			},
		},
		{
			name:    "wildcard dot",
			pattern: ".",
			want: func() *RegexNode {
				return &RegexNode{Type: NodeTypeGroup, Capturing: true, Children: []*RegexNode{
					NewCharGroupMatch(WildcardMatcher),
				}}
			},
		},
		{
			name:    "character class [abc]",
			pattern: "[abc]",
			want: func() *RegexNode {
				cg := &CharGroupMatcher{Chars: []byte{'a', 'b', 'c'}, Label: "abc"}
				return &RegexNode{Type: NodeTypeGroup, Capturing: true, Children: []*RegexNode{
					NewCharGroupMatch(cg),
				}}
			},
		},
		{
			name:    "negated character class [^abc]",
			pattern: "[^abc]",
			want: func() *RegexNode {
				cg := &CharGroupMatcher{Chars: []byte{'a', 'b', 'c'}, Negate: true, Label: "abc"}
				return &RegexNode{Type: NodeTypeGroup, Capturing: true, Children: []*RegexNode{
					NewCharGroupMatch(cg),
				}}
			},
		},
		{
			name:    "character class with \\d and literal",
			pattern: "[P\\d]",
			want: func() *RegexNode {
				cg := &CharGroupMatcher{Chars: []byte{'P'}, Ranges: [][2]byte{{'0', '9'}}, Negate: false, Label: "P\\d"}
				return &RegexNode{Type: NodeTypeGroup, Capturing: true, Children: []*RegexNode{
					NewCharGroupMatch(cg),
				}}
			},
		},
		{
			name:    "range character class [a-c]",
			pattern: "[a-c]",
			want: func() *RegexNode {
				cg := &CharGroupMatcher{Chars: []byte{}, Ranges: [][2]byte{{'a', 'c'}}, Label: "a-c"}
				return &RegexNode{Type: NodeTypeGroup, Capturing: true, Children: []*RegexNode{
					NewCharGroupMatch(cg),
				}}
			},
		},
		{
			name:    "anchors ^ab$",
			pattern: "^ab$",
			want: func() *RegexNode {
				return &RegexNode{Type: NodeTypeGroup, Capturing: true, Children: []*RegexNode{
					NewCaretAnchor(),
					NewLiteralMatch('a'),
					NewLiteralMatch('b'),
					NewDollarAnchor(),
				}}
			},
		},
		{
			name:    "quantifier plus a+",
			pattern: "a+",
			want: func() *RegexNode {
				n := NewLiteralMatch('a')
				n.Quantifier = QuantifierPlus
				return &RegexNode{Type: NodeTypeGroup, Capturing: true, Children: []*RegexNode{n}}
			},
		},
		{
			name:    "quantifier optional a?",
			pattern: "a?",
			want: func() *RegexNode {
				n := NewLiteralMatch('a')
				n.Quantifier = QuantifierOptional
				return &RegexNode{Type: NodeTypeGroup, Capturing: true, Children: []*RegexNode{n}}
			},
		},
		{
			name:    "simple capturing group (ab)",
			pattern: "(ab)",
			want: func() *RegexNode {
				g := NewGroup([]*RegexNode{NewLiteralMatch('a'), NewLiteralMatch('b')})
				g.Capturing = true
				return &RegexNode{Type: NodeTypeGroup, Capturing: true, Children: []*RegexNode{g}}
			},
		},
		{
			name:    "capturing group with plus (ab)+",
			pattern: "(ab)+",
			want: func() *RegexNode {
				g := NewGroup([]*RegexNode{NewLiteralMatch('a'), NewLiteralMatch('b')})
				g.Capturing = true
				g.Quantifier = QuantifierPlus
				return &RegexNode{Type: NodeTypeGroup, Capturing: true, Children: []*RegexNode{g}}
			},
		},
		{
			name:    "simple alternation a|b|c",
			pattern: "a|b|c",
			want: func() *RegexNode {
				alt := NewAlternation([]*RegexNode{NewLiteralMatch('a'), NewLiteralMatch('b'), NewLiteralMatch('c')})
				return &RegexNode{Type: NodeTypeGroup, Capturing: true, Children: []*RegexNode{alt}}
			},
		},
		{
			name:    "nested alternation ((ab)|c)+",
			pattern: "((ab)|c)+",
			want: func() *RegexNode {
				innerGroup := NewGroup([]*RegexNode{NewLiteralMatch('a'), NewLiteralMatch('b')})
				innerGroup.Capturing = true
				alt := NewAlternation([]*RegexNode{innerGroup, NewLiteralMatch('c')})
				alt.Capturing = true
				alt.Quantifier = QuantifierPlus
				return &RegexNode{Type: NodeTypeGroup, Capturing: true, Children: []*RegexNode{alt}}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.pattern).Parse()
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			expected := tt.want()
			if !nodesEqual(got, expected) {
				t.Errorf("Parse() AST mismatch for pattern %q\n got: %#v\nwant: %#v", tt.pattern, got, expected)
			}
		})
	}
}

func TestParser_Parse_Errors(t *testing.T) {
	tests := []string{
		"[abc", // unmatched [
		"(ab",  // unmatched (
		"\\",   // dangling escape
	}

	for _, pattern := range tests {
		t.Run(pattern, func(t *testing.T) {
			_, err := New(pattern).Parse()
			if err == nil {
				t.Fatalf("expected error for pattern %q, got nil", pattern)
			}
		})
	}
}
