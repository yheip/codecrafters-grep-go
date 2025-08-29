package regex

import (
	"fmt"
	"io"
)

// CompiledRegex represents a compiled regular expression as an NFA.
type CompiledRegex struct {
	initialState *State
	endingState  *State
}

func (re *CompiledRegex) SetInitialState(s *State) {
	re.initialState = s
}

func (re *CompiledRegex) SetEndingState(s *State) {
	re.endingState = s
}

func (re *CompiledRegex) InitialState() *State {
	return re.initialState
}

func (re *CompiledRegex) EndingState() *State {
	return re.endingState
}

// appendRegex connect the ending state of the current regex to the initial state of the other regex
func (re *CompiledRegex) appendRegex(other *CompiledRegex) {
	re.endingState.Append(other.initialState)
	re.endingState = other.endingState
}

type State struct {
	Transitions    []Transition
	StartingGroups []string
	EndingGroups   []string
}

func NewState() *State {
	return &State{
		Transitions: []Transition{},
	}
}

func (s *State) AddTransition(to *State, matcher Matcher) {
	s.Transitions = append(s.Transitions, Transition{
		To:      to,
		Matcher: matcher,
	})
}

func (s *State) PrependTransition(to *State, matcher Matcher) {
	s.Transitions = append([]Transition{{To: to, Matcher: matcher}}, s.Transitions...)
}

func (s *State) AddStartingGroup(name string) {
	s.StartingGroups = append(s.StartingGroups, name)
}

func (s *State) AddEndingGroup(name string) {
	s.EndingGroups = append(s.EndingGroups, name)
}

// Append merges another state's transitions and groups into the current state.
func (s *State) Append(other *State) {
	s.Transitions = append(s.Transitions, other.Transitions...)
	s.StartingGroups = append(s.StartingGroups, other.StartingGroups...)
	s.EndingGroups = append(s.EndingGroups, other.EndingGroups...)
}

type Stringer interface {
	String() string
}

type Transition struct {
	To *State
	Matcher
}

type Matcher interface {
	Match(b byte) bool
	IsEpsilon() bool
}

type CharMatcher struct {
	Char byte
}

func (m CharMatcher) Match(b byte) bool {
	return m.Char == b
}

func (m CharMatcher) IsEpsilon() bool {
	return false
}

func (m CharMatcher) String() string {
	return string(m.Char)
}

type EpsilonMatcher struct{}

func (m EpsilonMatcher) Match(b byte) bool {
	return true
}

func (m EpsilonMatcher) IsEpsilon() bool {
	return true
}

func (m EpsilonMatcher) String() string {
	return "ε"
}

func printRegex(w io.Writer, re *CompiledRegex) {
	idMap := make(map[*State]int)
	start := re.initialState
	assignIDs(start, idMap, new(int))
	fmt.Fprintf(w, "Start: %d, End: %d\nEdges:\n", idMap[start], idMap[re.endingState])
	visited := make(map[*State]bool)
	printEdges(w, start, visited, idMap)
}

func BuildIDMap(s *State) map[*State]int {
	idMap := make(map[*State]int)
	assignIDs(s, idMap, new(int))
	return idMap
}

func assignIDs(s *State, idMap map[*State]int, stateCounter *int) {
	if _, exists := idMap[s]; exists {
		return
	}

	idMap[s] = *stateCounter
	*stateCounter++

	for _, transition := range s.Transitions {
		assignIDs(transition.To, idMap, stateCounter)
	}
}

func printEdges(w io.Writer, s *State, visited map[*State]bool, idMap map[*State]int) {
	if visited[s] {
		return
	}

	visited[s] = true
	sourceID := idMap[s]

	for _, transition := range s.Transitions {
		targetID := idMap[transition.To]

		if t, ok := transition.Matcher.(Stringer); ok {
			fmt.Fprintf(w, "  %d --[%s]--> %d\n", sourceID, t.String(), targetID)
		} else {
			fmt.Fprintf(w, "  %d --[unknown]--> %d\n", sourceID, targetID)
		}
	}

	for _, transition := range s.Transitions {
		printEdges(w, transition.To, visited, idMap)
	}
}
