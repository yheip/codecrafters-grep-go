package matcher

import (
	"testing"

	"github.com/codecrafters-io/grep-starter-go/app/regex"
)

func TestMatch(t *testing.T) {
	type args struct {
		input []byte
		want  bool
	}

	tests := []struct {
		name string
		re   func() *regex.CompiledRegex
		args []args
	}{
		{
			name: "single char match",
			re: func() *regex.CompiledRegex {
				s0 := &regex.State{}
				s1 := &regex.State{}
				s0.AddTransition(s1, regex.CharMatcher{Char: 'a'})

				re := &regex.CompiledRegex{}
				re.SetInitialState(s0)
				re.SetEndingState(s1)

				return re
			},
			args: []args{
				{input: []byte("a"), want: true},
				{input: []byte("b"), want: false},
			},
		},
		{
			name: "with plus quantifier",
			re: func() *regex.CompiledRegex {
				s0 := &regex.State{}
				s1 := &regex.State{}
				s2 := &regex.State{}
				s0.AddTransition(s1, regex.CharMatcher{Char: 'a'})
				s1.AddTransition(s2, regex.CharMatcher{Char: 'b'})
				s2.AddTransition(s1, regex.EpsilonMatcher{})

				re := &regex.CompiledRegex{}
				re.SetInitialState(s0)
				re.SetEndingState(s2)

				return re
			},
			args: []args{
				{input: []byte("ab"), want: true},
				{input: []byte("abab"), want: true},
				{input: []byte("a"), want: false},
			},
		},
		{
			name: "simple alternation", // a|b|c
			re: func() *regex.CompiledRegex {
				s := make([]*regex.State, 5)
				for i := range s {
					s[i] = regex.NewState()
				}

				s[0].AddTransition(s[1], regex.CharMatcher{Char: 'a'})
				s[0].AddTransition(s[2], regex.CharMatcher{Char: 'b'})
				s[0].AddTransition(s[3], regex.CharMatcher{Char: 'c'})
				s[1].AddTransition(s[4], regex.EpsilonMatcher{})
				s[2].AddTransition(s[4], regex.EpsilonMatcher{})
				s[3].AddTransition(s[4], regex.EpsilonMatcher{})

				re := &regex.CompiledRegex{}
				re.SetInitialState(s[0])
				re.SetEndingState(s[4])

				return re
			},
			args: []args{
				{input: []byte("a"), want: true},
				{input: []byte("b"), want: true},
				{input: []byte("c"), want: true},
				{input: []byte("d"), want: false},
			},
		},
		{
			name: "simple group with plus quantifier", // (ab)+
			re: func() *regex.CompiledRegex {
				s := make([]*regex.State, 5)
				for i := range s {
					s[i] = regex.NewState()
				}
				s[0].AddTransition(s[1], regex.EpsilonMatcher{})
				s[1].AddTransition(s[2], regex.CharMatcher{Char: 'a'})
				s[2].AddTransition(s[3], regex.CharMatcher{Char: 'b'})
				s[3].AddTransition(s[1], regex.EpsilonMatcher{}) // loop back to s1
				s[3].AddTransition(s[4], regex.EpsilonMatcher{})

				re := &regex.CompiledRegex{}
				re.SetInitialState(s[0])
				re.SetEndingState(s[4])

				return re
			},
			args: []args{
				{input: []byte("ab"), want: true},
				{input: []byte("abab"), want: true},
				{input: []byte("cabab"), want: true},
				{input: []byte("abc"), want: true},
				{input: []byte("xabc"), want: true},
				{input: []byte("acb"), want: false},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re := tt.re()
			for _, arg := range tt.args {
				t.Run(string(arg.input), func(t *testing.T) {
					if got := Match(arg.input, re); got != arg.want {
						t.Errorf("Match() = %v, want %v", got, arg.want)
					}
				})
			}
		})
	}
}
