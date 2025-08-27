package matcher

import (
	"maps"
	"slices"

	"github.com/codecrafters-io/grep-starter-go/app/regex"
)

type searchState struct {
	idx            int
	state          *regex.State
	epsilonVisited map[*regex.State]bool // To avoid infinite loops on epsilon transitions
}

func Match(input []byte, re *regex.CompiledRegex) bool {
	for i := 0; i <= len(input); i++ {
		if matchAt(i, input, re) {
			return true
		}
	}

	return false
}

func matchAt(i int, input []byte, re *regex.CompiledRegex) bool {
	stack := []searchState{{i, re.InitialState(), map[*regex.State]bool{}}}

	for len(stack) > 0 {
		current := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if current.state == re.EndingState() {
			return true
		}

		// Go through transitions in reverse order to maintain the original order when using a stack
		for _, tr := range slices.Backward(current.state.Transitions) {
			if tr.IsEpsilon() {
				epilonVisited := maps.Clone(current.epsilonVisited)
				//.Don't consume input on epsilon transitions
				if epilonVisited[tr.To] {
					continue // Avoid infinite loops on epsilon transitions
				}

				epilonVisited[tr.To] = true
				stack = append(stack, searchState{current.idx, tr.To, epilonVisited})
			} else if current.idx < len(input) && tr.Match(input[current.idx]) {
				epilonVisited := map[*regex.State]bool{} // Reset epsilon visited on non-epsilon transitions

				stack = append(stack, searchState{current.idx + 1, tr.To, epilonVisited})
			}
		}
	}

	return false
}
