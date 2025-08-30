package regex

import (
	"slices"
	"testing"
)

func Test_regexEqual(t *testing.T) {
	tests := []struct {
		name  string
		left  func() *CompiledRegex
		right func() *CompiledRegex
		want  bool
	}{
		{
			name: "same transition states",
			left: func() *CompiledRegex {
				s0 := &State{}
				s1 := &State{}
				s0.AddTransition(s1, CharMatcher{Char: 'a'})
				return &CompiledRegex{initialState: s0, endingState: s1}
			},
			right: func() *CompiledRegex {
				s0 := &State{}
				s1 := &State{}
				s0.AddTransition(s1, CharMatcher{Char: 'a'})
				return &CompiledRegex{initialState: s0, endingState: s1}
			},
			want: true,
		},
		{
			name: "deeper with epsilon transitions",
			left: func() *CompiledRegex {
				s0 := &State{}
				s1 := &State{}
				s2 := &State{}
				s3 := &State{}
				s0.AddTransition(s1, EpsilonMatcher{})
				s1.AddTransition(s2, CharMatcher{Char: 'a'})
				s2.AddTransition(s3, CharMatcher{Char: 'b'})
				s2.AddTransition(s1, EpsilonMatcher{}) // loop back to s1

				return &CompiledRegex{initialState: s0, endingState: s3}
			},
			right: func() *CompiledRegex {
				s0 := &State{}
				s1 := &State{}
				s2 := &State{}
				s3 := &State{}
				s0.AddTransition(s1, EpsilonMatcher{})
				s1.AddTransition(s2, CharMatcher{Char: 'a'})
				s2.AddTransition(s3, CharMatcher{Char: 'b'})
				s2.AddTransition(s1, EpsilonMatcher{}) // loop back to s1

				return &CompiledRegex{initialState: s0, endingState: s3}
			},
			want: true,
		},
		{
			name: "with infinite epsilon loop",
			left: func() *CompiledRegex {
				s0 := &State{}
				s1 := &State{}
				s2 := &State{}
				s0.AddTransition(s1, CharMatcher{Char: 'a'})
				s1.AddTransition(s1, EpsilonMatcher{})
				s1.AddTransition(s2, CharMatcher{Char: 'b'})

				return &CompiledRegex{initialState: s0, endingState: s2}
			},
			right: func() *CompiledRegex {
				s0 := &State{}
				s1 := &State{}
				s2 := &State{}
				s0.AddTransition(s1, CharMatcher{Char: 'a'})
				s1.AddTransition(s1, EpsilonMatcher{})
				s1.AddTransition(s2, CharMatcher{Char: 'b'})

				return &CompiledRegex{initialState: s0, endingState: s2}
			},
			want: true,
		},
		{
			name: "different transition states",
			left: func() *CompiledRegex {
				s0 := &State{}
				s1 := &State{}
				s0.AddTransition(s1, CharMatcher{Char: 'a'})
				return &CompiledRegex{initialState: s0, endingState: s1}
			},
			right: func() *CompiledRegex {
				s0 := &State{}
				s1 := &State{}
				s0.AddTransition(s1, CharMatcher{Char: 'b'})
				return &CompiledRegex{initialState: s0, endingState: s1}
			},
			want: false,
		},
		{
			name: "different structure",
			left: func() *CompiledRegex {
				s0 := &State{}
				s1 := &State{}
				s2 := &State{}
				s0.AddTransition(s1, CharMatcher{Char: 'a'})
				s1.AddTransition(s2, CharMatcher{Char: 'b'})
				return &CompiledRegex{initialState: s0, endingState: s2}
			},
			right: func() *CompiledRegex {
				s0 := &State{}
				s1 := &State{}
				s0.AddTransition(s1, CharMatcher{Char: 'a'})
				s0.AddTransition(s1, CharMatcher{Char: 'b'})
				return &CompiledRegex{initialState: s0, endingState: s1}
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			left, right := tt.left(), tt.right()

			if got := regexsEqual(left, right); got != tt.want {
				t.Errorf("CompiledRegex.Equal() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Equal deeply compares two CompiledRegex instances for equality.
func regexsEqual(c1, c2 *CompiledRegex) bool {
	// Quick check for pointer equality
	if c1 == c2 {
		return true
	}

	// Check if both initial and ending states are the same (pointer equality)
	if c1.initialState == c2.initialState && c1.endingState == c2.endingState {
		return true
	}

	return statesEqual(c1.initialState, c2.initialState) && statesEqual(c1.endingState, c2.endingState)
}

func statesEqual(s1, s2 *State) bool {
	if s1 == nil && s2 == nil {
		return true
	}
	if s1 == nil || s2 == nil {
		return false
	}

	visited := make(map[*State]*State)
	return statesEqualHelper(s1, s2, visited)
}

func statesEqualHelper(s1, s2 *State, visited map[*State]*State) bool {
	if s1 == s2 {
		return true
	}

	if mappedState, exists := visited[s1]; exists {
		return mappedState == s2
	}

	visited[s1] = s2

	if !slices.Equal(
		slices.Sorted(slices.Values(s1.StartingGroups)),
		slices.Sorted(slices.Values(s2.StartingGroups)),
	) || !slices.Equal(
		slices.Sorted(slices.Values(s1.EndingGroups)),
		slices.Sorted(slices.Values(s2.EndingGroups)),
	) {
		return false
	}

	if len(s1.Transitions) != len(s2.Transitions) {
		return false
	}

	for i, transition := range s1.Transitions {
		otherTransition := s2.Transitions[i]

		if !matchersEqual(transition.Matcher, otherTransition.Matcher) {
			return false
		}

		if !statesEqualHelper(transition.To, otherTransition.To, visited) {
			return false
		}
	}

	return true
}

func matchersEqual(m1, m2 Matcher) bool {
	if m1 == nil && m2 == nil {
		return true
	}
	if m1 == nil || m2 == nil {
		return false
	}

	switch v1 := m1.(type) {
	case CharMatcher:
		if v2, ok := m2.(CharMatcher); ok {
			return v1.Char == v2.Char
		}
	case EpsilonMatcher:
		_, ok := m2.(EpsilonMatcher)
		return ok
	}

	return false
}
