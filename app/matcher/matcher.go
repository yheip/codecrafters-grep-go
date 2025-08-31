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
	groups         map[string]GroupMatch // Currently open groups
	capturedGroups map[string]GroupMatch // All captured groups so far
}

type GroupMatch struct {
	start int
	end   int
}

func (g GroupMatch) Start() int {
	return g.start
}

func (g GroupMatch) End() int {
	return g.end
}

type MatchArg struct {
	input          []byte
	pos            int
	capturedGroups map[string]GroupMatch
}

func (m MatchArg) Input() []byte {
	return m.input
}

func (m MatchArg) Pos() int {
	return m.pos
}

func (m MatchArg) Backreference(name string) (regex.GroupSpan, bool) {
	if match, exists := m.capturedGroups[name]; exists {
		return match, true
	}

	return nil, false
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
	idsmap := regex.BuildIDMap(re.InitialState())
	slog.Debug("Target State", "id", idsmap[re.EndingState()])
	for i := 0; i <= len(input); i++ {
		if matchedGrp := matchAt(i, input, re); matchedGrp != nil {
			// Convert GroupMatch to map[string]string
			slog.Debug("matchAt", "grp", matchedGrp)
			result := make(map[string]string)
			for name, match := range matchedGrp {
				if match.end != -1 { // Ensure the group was closed
					// Slice the input to get the matched substring
					result[name] = string(input[match.start:match.end])
				}
			}

			return result
		}
	}

	return nil
}

func matchAt(i int, input []byte, re *regex.CompiledRegex) map[string]GroupMatch {
	stack := []searchState{{
		idx:            i,
		state:          re.InitialState(),
		epsilonVisited: map[*regex.State]bool{},
		groups:         map[string]GroupMatch{},
		capturedGroups: map[string]GroupMatch{},
	}}

	idsmap := regex.BuildIDMap(re.InitialState())

	for len(stack) > 0 {
		current := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		for _, grp := range current.state.StartingGroups {
			current.groups[grp] = GroupMatch{current.idx, -1}
		}
		for _, grp := range current.state.EndingGroups {
			if match, exists := current.groups[grp]; exists {
				match.end = current.idx
				current.groups[grp] = match
				current.capturedGroups[grp] = match
			}
		}
		slog.Debug("At", "state", idsmap[current.state], "idx", current.idx, "groups", current.groups)

		if current.state == re.EndingState() {
			return current.capturedGroups
		}

		// Go through transitions in reverse order to maintain the original order when using a stack
		for _, tr := range slices.Backward(current.state.Transitions) {
			groups := maps.Clone(current.groups) // Clone matched groups for each transition
			capturedGroups := maps.Clone(current.capturedGroups)

			arg := MatchArg{
				input:          input,
				pos:            current.idx,
				capturedGroups: capturedGroups,
			}
			if n, ok := tr.Match(arg); ok {
				if n > 0 { // Non-epsilon transition
					// Reset epsilon visited on non-epsilon transitions
					epilonVisited := map[*regex.State]bool{}
					stack = append(stack, searchState{current.idx + n, tr.To, epilonVisited, groups, capturedGroups})

					continue

				}

				// Epsilon transition
				epilonVisited := maps.Clone(current.epsilonVisited)
				if epilonVisited[tr.To] {
					continue // Avoid infinite loops on epsilon transitions
				}

				epilonVisited[tr.To] = true
				//.Don't consume input on epsilon transitions
				stack = append(stack, searchState{current.idx, tr.To, epilonVisited, groups, capturedGroups})
			}
		}
	}

	return nil
}
