package regex

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/codecrafters-io/grep-starter-go/app/parser"
)

func TestCompile(t *testing.T) {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	tests := []struct {
		name string
		root *parser.RegexNode
		want func() *CompiledRegex
	}{
		{
			name: "single char", // a
			root: &parser.RegexNode{
				Type: parser.NodeTypeGroup,
				Children: []*parser.RegexNode{
					parser.NewLiteral('a'),
				},
			},
			want: func() *CompiledRegex {
				return NewSingleCharRegex('a')
			},
		},
		{
			name: "multiple chars", // abc
			root: &parser.RegexNode{
				Type: parser.NodeTypeGroup,
				Children: []*parser.RegexNode{
					parser.NewLiteral('a'),
					parser.NewLiteral('b'),
					parser.NewLiteral('c'),
				},
			},
			want: func() *CompiledRegex {
				s0 := NewState()
				s1 := NewState()
				s2 := NewState()
				s3 := NewState()
				s0.AddTransition(s1, CharMatcher{Char: 'a'})
				s1.AddTransition(s2, CharMatcher{Char: 'b'})
				s2.AddTransition(s3, CharMatcher{Char: 'c'})

				return &CompiledRegex{initialState: s0, endingState: s3}
			},
		},
		{
			name: "single char with plus quantifier", // a+
			root: &parser.RegexNode{
				Type: parser.NodeTypeGroup,
				Children: []*parser.RegexNode{
					parser.NewLiteral('a').WithQuantifier(parser.QuantifierPlus),
				},
			},
			want: func() *CompiledRegex {
				s0 := NewState()
				s1 := NewState()
				s2 := NewState()
				s3 := NewState()
				s0.AddTransition(s1, EpsilonMatcher{})
				s1.AddTransition(s2, CharMatcher{Char: 'a'})
				s2.AddTransition(s1, EpsilonMatcher{}) // loop back to s1
				s2.AddTransition(s3, EpsilonMatcher{})

				return &CompiledRegex{initialState: s0, endingState: s3}
			},
		},
		{
			name: "single char with optional quantifier", // a?
			root: &parser.RegexNode{
				Type: parser.NodeTypeGroup,
				Children: []*parser.RegexNode{
					parser.NewLiteral('a').WithQuantifier(parser.QuantifierOptional),
				},
			},
			want: func() *CompiledRegex {
				s0 := NewState()
				s1 := NewState()
				s0.AddTransition(s1, CharMatcher{Char: 'a'})
				s0.AddTransition(s1, EpsilonMatcher{})

				return &CompiledRegex{initialState: s0, endingState: s1}
			},
		},
		{
			name: "simple alternation", // a|b|c
			root: &parser.RegexNode{
				Type: parser.NodeTypeGroup,
				Children: []*parser.RegexNode{
					parser.NewAlternation([]*parser.RegexNode{
						parser.NewLiteral('a'),
						parser.NewLiteral('b'),
						parser.NewLiteral('c'),
					}),
				},
			},
			want: func() *CompiledRegex {
				s := make([]*State, 8)
				for i := range s {
					s[i] = NewState()
				}

				s[0].AddTransition(s[1], EpsilonMatcher{})
				s[1].AddTransition(s[2], CharMatcher{Char: 'a'})
				s[2].AddTransition(s[3], EpsilonMatcher{})
				s[0].AddTransition(s[4], EpsilonMatcher{})
				s[4].AddTransition(s[5], CharMatcher{Char: 'b'})
				s[5].AddTransition(s[3], EpsilonMatcher{})
				s[0].AddTransition(s[6], EpsilonMatcher{})
				s[6].AddTransition(s[7], CharMatcher{Char: 'c'})
				s[7].AddTransition(s[3], EpsilonMatcher{})

				return &CompiledRegex{initialState: s[0], endingState: s[3]}
			},
		},
		{
			name: "simple group", // (ab)
			root: &parser.RegexNode{
				Type: parser.NodeTypeGroup,
				Children: []*parser.RegexNode{
					parser.NewLiteral('a'),
					parser.NewLiteral('b'),
				},
			},
			want: func() *CompiledRegex {
				s0 := NewState()
				s1 := NewState()
				s2 := NewState()
				s0.AddTransition(s1, CharMatcher{Char: 'a'})
				s1.AddTransition(s2, CharMatcher{Char: 'b'})

				return &CompiledRegex{initialState: s0, endingState: s2}
			},
		},
		{
			name: "simple group with capturing", // (ab)
			root: &parser.RegexNode{
				Type: parser.NodeTypeGroup,
				Children: []*parser.RegexNode{
					{
						Type: parser.NodeTypeGroup,
						Children: []*parser.RegexNode{
							parser.NewLiteral('a'),
							parser.NewLiteral('b'),
						},
						Capturing: true,
					},
				},
				Capturing: true,
			},
			want: func() *CompiledRegex {
				s0 := NewState()
				s1 := NewState()
				s2 := NewState()
				s0.AddStartingGroup("0")
				s0.AddStartingGroup("1")
				s0.AddTransition(s1, CharMatcher{Char: 'a'})
				s1.AddTransition(s2, CharMatcher{Char: 'b'})
				s2.AddEndingGroup("1")
				s2.AddEndingGroup("0")

				return &CompiledRegex{initialState: s0, endingState: s2}
			},
		},
		{
			name: "simple group with plus quantifier", // (ab)+
			root: &parser.RegexNode{
				Type: parser.NodeTypeGroup,
				Children: []*parser.RegexNode{
					{
						Type: parser.NodeTypeGroup,
						Children: []*parser.RegexNode{
							parser.NewLiteral('a'),
							parser.NewLiteral('b'),
						},
						Quantifier: parser.QuantifierPlus,
					},
				},
			},
			want: func() *CompiledRegex {
				s := make([]*State, 5)
				for i := range s {
					s[i] = NewState()
				}
				s[0].AddTransition(s[1], EpsilonMatcher{})
				s[1].AddTransition(s[2], CharMatcher{Char: 'a'})
				s[2].AddTransition(s[3], CharMatcher{Char: 'b'})
				s[3].AddTransition(s[1], EpsilonMatcher{}) // loop back to s1
				s[3].AddTransition(s[4], EpsilonMatcher{})

				return &CompiledRegex{initialState: s[0], endingState: s[4]}
			},
		},
		{
			name: "simple group with capturing and plus quantifier", // (ab)+
			root: &parser.RegexNode{
				Type: parser.NodeTypeGroup,
				Children: []*parser.RegexNode{
					{
						Type: parser.NodeTypeGroup,
						Children: []*parser.RegexNode{
							parser.NewLiteral('a'),
							parser.NewLiteral('b'),
						},
						Quantifier: parser.QuantifierPlus,
						Capturing:  true,
					},
				},
				Capturing: true,
			},
			want: func() *CompiledRegex {
				s := make([]*State, 5)
				for i := range s {
					s[i] = NewState()
				}
				s[0].AddTransition(s[1], EpsilonMatcher{})
				s[0].AddStartingGroup("0")
				s[1].AddStartingGroup("1")
				s[1].AddTransition(s[2], CharMatcher{Char: 'a'})
				s[2].AddTransition(s[3], CharMatcher{Char: 'b'})
				s[3].AddTransition(s[1], EpsilonMatcher{}) // loop back to s1
				s[3].AddEndingGroup("1")
				s[3].AddTransition(s[4], EpsilonMatcher{})
				s[4].AddEndingGroup("0")

				return &CompiledRegex{initialState: s[0], endingState: s[4]}
			},
		},
		{
			name: "nested group with alternation and capturing", // ((ab)|c)+
			root: &parser.RegexNode{
				Type: parser.NodeTypeGroup,
				Children: []*parser.RegexNode{
					{
						Type: parser.NodeTypeAlternation,
						Children: []*parser.RegexNode{
							{
								Type: parser.NodeTypeGroup,
								Children: []*parser.RegexNode{
									parser.NewLiteral('a'),
									parser.NewLiteral('b'),
								},
								Capturing: true,
							},
							parser.NewLiteral('c'),
						},
						Quantifier: parser.QuantifierPlus,
						Capturing:  true,
					},
				},
				Capturing: true,
			},
			want: func() *CompiledRegex {
				s := make([]*State, 9)
				for i := range s {
					s[i] = NewState()
				}
				s[0].AddStartingGroup("0")
				s[0].AddTransition(s[1], EpsilonMatcher{})
				s[1].AddTransition(s[2], EpsilonMatcher{})
				s[1].AddStartingGroup("1")
				s[2].AddStartingGroup("2")
				s[2].AddTransition(s[3], CharMatcher{Char: 'a'})
				s[3].AddTransition(s[4], CharMatcher{Char: 'b'})
				s[4].AddTransition(s[5], EpsilonMatcher{})
				s[4].AddEndingGroup("2")
				s[5].AddTransition(s[1], EpsilonMatcher{})
				s[5].AddTransition(s[6], EpsilonMatcher{})
				s[1].AddTransition(s[7], EpsilonMatcher{})
				s[7].AddTransition(s[8], CharMatcher{Char: 'c'})
				s[8].AddTransition(s[5], EpsilonMatcher{})
				s[5].AddEndingGroup("1")
				s[6].AddEndingGroup("0")

				return &CompiledRegex{initialState: s[0], endingState: s[6]}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expected := tt.want()
			if got, err := Compile(tt.root); err != nil {
				t.Errorf("Compile() error = %v", err)
			} else if !regexsEqual(got, expected) {
				gotBuf, expectedBuf := &bytes.Buffer{}, &bytes.Buffer{}
				printRegex(expectedBuf, expected)
				printRegex(gotBuf, got)

				t.Errorf("Compile() = \n%v\nwant = \n%v", gotBuf.String(), expectedBuf.String())
			}
		})
	}

}
