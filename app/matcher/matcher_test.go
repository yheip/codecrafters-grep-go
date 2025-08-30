package matcher

import (
	"log/slog"
	"testing"

	"github.com/codecrafters-io/grep-starter-go/app/parser"
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
				s0.AddTransition(s1, literalCharTransitioner('a'))

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
				s0.AddTransition(s1, literalCharTransitioner('a'))
				s1.AddTransition(s2, literalCharTransitioner('b'))
				s2.AddTransition(s1, regex.EpsilonTransitioner{})

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

				s[0].AddTransition(s[1], literalCharTransitioner('a'))
				s[0].AddTransition(s[2], literalCharTransitioner('b'))
				s[0].AddTransition(s[3], literalCharTransitioner('c'))
				s[1].AddTransition(s[4], regex.EpsilonTransitioner{})
				s[2].AddTransition(s[4], regex.EpsilonTransitioner{})
				s[3].AddTransition(s[4], regex.EpsilonTransitioner{})

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
				s[0].AddTransition(s[1], regex.EpsilonTransitioner{})
				s[1].AddTransition(s[2], literalCharTransitioner('a'))
				s[2].AddTransition(s[3], literalCharTransitioner('b'))
				s[3].AddTransition(s[1], regex.EpsilonTransitioner{}) // loop back to s1
				s[3].AddTransition(s[4], regex.EpsilonTransitioner{})

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
		{
			name: "nested group with alternation", // ((ab)|c)+
			re: func() *regex.CompiledRegex {
				s := make([]*regex.State, 7)
				for i := range s {
					s[i] = regex.NewState()
				}
				s[0].AddTransition(s[1], regex.EpsilonTransitioner{})
				s[1].AddTransition(s[2], literalCharTransitioner('a'))
				s[2].AddTransition(s[3], literalCharTransitioner('b'))
				s[1].AddTransition(s[4], literalCharTransitioner('c'))
				s[3].AddTransition(s[5], regex.EpsilonTransitioner{})
				s[4].AddTransition(s[5], regex.EpsilonTransitioner{})
				s[5].AddTransition(s[1], regex.EpsilonTransitioner{})
				s[5].AddTransition(s[6], regex.EpsilonTransitioner{})

				re := &regex.CompiledRegex{}
				re.SetInitialState(s[0])
				re.SetEndingState(s[6])

				return re
			},
			args: []args{
				{input: []byte("ab"), want: true},
				{input: []byte("c"), want: true},
				{input: []byte("abc"), want: true},
				{input: []byte("abccab"), want: true},
				{input: []byte("a"), want: false},
				{input: []byte("ac"), want: true},
				{input: []byte("cabx"), want: true},
				{input: []byte("x"), want: false},
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

func TestMatchWithCaptureGroups(t *testing.T) {
	type args struct {
		input  []byte
		want   bool
		groups map[string]string
	}

	slog.SetLogLoggerLevel(slog.LevelDebug)

	tests := []struct {
		name string
		re   func() *regex.CompiledRegex
		args []args
	}{
		{
			name: "single capturing group",
			re: func() *regex.CompiledRegex {
				s0 := &regex.State{}
				s1 := &regex.State{}
				s2 := &regex.State{}
				s0.AddTransition(s1, literalCharTransitioner('a'))
				s1.AddTransition(s2, literalCharTransitioner('b'))
				s0.AddStartingGroup("0")
				s0.AddStartingGroup("1")
				s2.AddEndingGroup("0")
				s2.AddEndingGroup("1")

				re := &regex.CompiledRegex{}
				re.SetInitialState(s0)
				re.SetEndingState(s2)

				return re
			},
			args: []args{
				{input: []byte("ab"), want: true, groups: map[string]string{"0": "ab", "1": "ab"}},
				{input: []byte("a"), want: false, groups: map[string]string{}},
			},
		},
		{
			name: "simple group with plus quantifier", // (ab)+
			re: func() *regex.CompiledRegex {
				s := make([]*regex.State, 5)
				for i := range s {
					s[i] = regex.NewState()
				}
				s[0].AddStartingGroup("0")
				s[0].AddTransition(s[1], regex.EpsilonTransitioner{})
				s[1].AddTransition(s[2], literalCharTransitioner('a'))
				s[1].AddStartingGroup("1")
				s[2].AddTransition(s[3], literalCharTransitioner('b'))
				s[3].AddTransition(s[1], regex.EpsilonTransitioner{}) // loop back to s1
				s[3].AddTransition(s[4], regex.EpsilonTransitioner{})
				s[3].AddEndingGroup("1")
				s[4].AddEndingGroup("0")

				re := &regex.CompiledRegex{}
				re.SetInitialState(s[0])
				re.SetEndingState(s[4])

				return re
			},
			args: []args{
				{input: []byte("ab"), want: true, groups: map[string]string{"0": "ab", "1": "ab"}},
				{input: []byte("abab"), want: true, groups: map[string]string{"0": "abab", "1": "ab"}},
			},
		},
		{
			name: "nested group with alternation", // ((ab)|c)+
			re: func() *regex.CompiledRegex {
				s := make([]*regex.State, 9)
				for i := range s {
					s[i] = regex.NewState()
				}
				s[0].AddStartingGroup("0")
				s[0].AddTransition(s[1], regex.EpsilonTransitioner{})
				s[1].AddTransition(s[2], regex.EpsilonTransitioner{})
				s[1].AddStartingGroup("1")
				s[2].AddStartingGroup("2")
				s[2].AddTransition(s[3], literalCharTransitioner('a'))
				s[3].AddTransition(s[4], literalCharTransitioner('b'))
				s[4].AddTransition(s[5], regex.EpsilonTransitioner{})
				s[4].AddEndingGroup("2")
				s[5].AddTransition(s[1], regex.EpsilonTransitioner{})
				s[5].AddTransition(s[6], regex.EpsilonTransitioner{})
				s[1].AddTransition(s[7], regex.EpsilonTransitioner{})
				s[7].AddTransition(s[8], literalCharTransitioner('c'))
				s[8].AddTransition(s[5], regex.EpsilonTransitioner{})
				s[5].AddEndingGroup("1")
				s[6].AddEndingGroup("0")

				re := &regex.CompiledRegex{}
				re.SetInitialState(s[0])
				re.SetEndingState(s[6])

				return re
			},
			args: []args{
				{input: []byte("ab"), want: true, groups: map[string]string{"0": "ab", "1": "ab", "2": "ab"}},

				{input: []byte("ababc"), want: true, groups: map[string]string{"0": "ababc", "1": "c", "2": "ab"}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re := tt.re()
			for _, arg := range tt.args {
				t.Run(string(arg.input), func(t *testing.T) {
					groups := MatchWithCaptureGroups(arg.input, re)

					if (groups != nil) != arg.want {
						t.Errorf("Match() = %v, want %v", groups != nil, arg.want)
						return
					}

					if len(groups) != len(arg.groups) {
						t.Errorf("Match() groups = %v, want %v", groups, arg.groups)
					}

					for k, v := range arg.groups {
						if groups[k] != v {
							t.Errorf("Match() groups[%s] = %v, want %v", k, groups[k], v)
						}
					}
				})
			}
		})
	}
}

func literalCharTransitioner(b byte) regex.Transitioner {
	return regex.CharTransitioner{Matcher: &parser.LiteralMatcher{Char: b}}
}
