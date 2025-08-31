package matcher

import (
	"log/slog"
	"maps"
	"slices"

	"github.com/codecrafters-io/grep-starter-go/app/regex"
)

type searchState struct {
	idx            int
	state          *regex.State
	epsilonVisited map[*regex.State]bool // To avoid infinite loops on epsilon transitions
	matchedGroups  map[string]GroupMatch
}

type GroupMatch struct {
	Start int
	End   int
}

func Match(input []byte, re *regex.CompiledRegex) bool {
	for i := 0; i <= len(input); i++ {
		if matchedGrp := matchAt(i, input, re); matchedGrp != nil {
			return true
		}
	}

	return false
}

func MatchWithCaptureGroups(input []byte, re *regex.CompiledRegex) map[string]string {
	for i := 0; i <= len(input); i++ {
		if matchedGrp := matchAt(i, input, re); matchedGrp != nil {
			// Convert GroupMatch to map[string]string
			slog.Debug("matchAt", "grp", matchedGrp)
			result := make(map[string]string)
			for name, match := range matchedGrp {
				if match.End != -1 { // Ensure the group was closed
					// Slice the input to get the matched substring
					result[name] = string(input[match.Start:match.End])
				}
			}

			return result
		}
	}

	return nil
}

func matchAt(i int, input []byte, re *regex.CompiledRegex) map[string]GroupMatch {
	stack := []searchState{{
		i, re.InitialState(), map[*regex.State]bool{}, map[string]GroupMatch{}},
	}

	idsmap := regex.BuildIDMap(re.InitialState())
	slog.Debug("Target State", "id", idsmap[re.EndingState()])

	for len(stack) > 0 {
		current := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		for _, grp := range current.state.StartingGroups {
			current.matchedGroups[grp] = GroupMatch{current.idx, -1}
		}
		for _, grp := range current.state.EndingGroups {
			if match, exists := current.matchedGroups[grp]; exists {
				match.End = current.idx
				current.matchedGroups[grp] = match
			}
		}
		slog.Debug("At", "state", idsmap[current.state], "idx", current.idx, "matchedGroups", current.matchedGroups)

		if current.state == re.EndingState() {
			return current.matchedGroups
		}

		// Go through transitions in reverse order to maintain the original order when using a stack
		for _, tr := range slices.Backward(current.state.Transitions) {
			matchedGroups := maps.Clone(current.matchedGroups) // Clone matched groups for each transition
			if tr.Match(input, current.idx) {
				if tr.Consumable() {
					epilonVisited := map[*regex.State]bool{} // Reset epsilon visited on non-epsilon transitions
					stack = append(stack, searchState{current.idx + 1, tr.To, epilonVisited, matchedGroups})

					continue

				}

				epilonVisited := maps.Clone(current.epsilonVisited)
				//.Don't consume input on epsilon transitions
				if epilonVisited[tr.To] {
					continue // Avoid infinite loops on epsilon transitions
				}

				epilonVisited[tr.To] = true
				stack = append(stack, searchState{current.idx, tr.To, epilonVisited, matchedGroups})
			}
		}
	}

	return nil
}
