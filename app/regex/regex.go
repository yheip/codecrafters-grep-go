package regex

import (
	"fmt"
	"io"
	"log/slog"
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

func (s *State) AddTransition(to *State, transitioner Transitioner) {
	s.Transitions = append(s.Transitions, Transition{
		To:           to,
		Transitioner: transitioner,
	})
}

func (s *State) PrependTransition(to *State, transitioner Transitioner) {
	s.Transitions = append([]Transition{{To: to, Transitioner: transitioner}}, s.Transitions...)
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
	Transitioner
}

type MatchArg interface {
	Pos() int
	Input() []byte
	Backreference(name string) (GroupSpan, bool)
}

type GroupSpan interface {
	Start() int
	End() int
}

// Transitioner defines the interface for transition conditions in the NFA.
type Transitioner interface {
	Match(arg MatchArg) (int, bool)
	Stringer
}

// Matcher defines the interface for matching a byte.
type Matcher interface {
	Match(b byte) bool
	Stringer
}

// CharTransitioner is a transition that matches a specific character.
type CharTransitioner struct {
	Matcher
}

func (m CharTransitioner) Match(arg MatchArg) (int, bool) {
	// Ensure position is within bounds
	input, pos := arg.Input(), arg.Pos()

	if pos >= len(input) {
		return 0, false
	}

	return 1, m.Matcher.Match(input[pos])
}

func (m CharTransitioner) String() string {
	return m.Matcher.String()
}

// EpsilonTransitioner is a transition that represents an epsilon transition.
// It always matches without consuming any input.
type EpsilonTransitioner struct{}

func (m EpsilonTransitioner) Match(arg MatchArg) (int, bool) {
	return 0, true
}

func (m EpsilonTransitioner) String() string {
	return "ε"
}

type EndOfStringTransitioner struct{}

func (m EndOfStringTransitioner) Match(arg MatchArg) (int, bool) {
	return 0, arg.Pos() >= len(arg.Input())
}

func (m EndOfStringTransitioner) String() string {
	return "$"
}

type StartOfStringTransitioner struct{}

func (m StartOfStringTransitioner) Match(arg MatchArg) (int, bool) {
	return 0, arg.Pos() == 0
}

func (m StartOfStringTransitioner) String() string {
	return "^"
}

type BackReferenceTransitioner struct {
	GroupName string
}

func (m BackReferenceTransitioner) Match(arg MatchArg) (int, bool) {
	slog.Debug("BackReferenceTransitioner", "GroupName", m.GroupName)
	match, exists := arg.Backreference(m.GroupName)
	if !exists || match.Start() == -1 || match.End() == -1 {
		return 0, false
	}
	slog.Debug("BackReferenceTransitioner", "match", match, "input", string(arg.Input()), "pos", arg.Pos())

	length := match.End() - match.Start()
	input, pos := arg.Input(), arg.Pos()

	if pos+length > len(input) {
		return 0, false
	}

	for i := range length {
		if input[match.Start()+i] != input[pos+i] {
			return 0, false
		}
	}

	return length, true
}

func (m BackReferenceTransitioner) String() string {
	return fmt.Sprintf(`\%s`, m.GroupName)
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

	fmt.Fprintf(w, "Groups at state %d: start=%v, end=%v\n", sourceID, s.StartingGroups, s.EndingGroups)

	for _, tr := range s.Transitions {
		targetID := idMap[tr.To]

		fmt.Fprintf(w, "  %d --[%s]--> %d\n", sourceID, tr.String(), targetID)
	}

	for _, transition := range s.Transitions {
		printEdges(w, transition.To, visited, idMap)
	}
}
