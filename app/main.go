package main

import (
	"fmt"
	"io"
	"os"
)

// Usage: echo <input_text> | your_program.sh -E <pattern>
func main() {
	if len(os.Args) < 3 || os.Args[1] != "-E" {
		fmt.Fprintf(os.Stderr, "usage: mygrep -E <pattern>\n")
		os.Exit(2) // 1 means no lines were selected, >1 means error
	}

	pattern := os.Args[2]

	line, err := io.ReadAll(os.Stdin) // assume we're only dealing with a single line
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: read input text: %v\n", err)
		os.Exit(2)
	}

	ok, err := matchLine(line, pattern)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(2)
	}

	if !ok {
		os.Exit(1)
	}

	// default exit code is 0 which means success
}

type Regex struct {
	matchStart bool
	tokens     []Class
}

func NewRegex() *Regex {
	return &Regex{
		tokens: []Class{},
	}
}

func (r *Regex) String() string {
	var result string
	if r.matchStart {
		result += "^"
	}

	for _, token := range r.tokens {
		result += token.String()
	}

	return result
}

func match(regex *Regex, line []byte) bool {
	var i int

	if regex.matchStart {
		return matchHere(regex.tokens, line)
	}

	for i < len(line) {
		if matchHere(regex.tokens, line[i:]) {
			return true
		}

		i++
	}

	return false
}

func matchHere(regex []Class, line []byte) bool {
	if len(regex) == 0 {
		return true // empty regex matches everything
	}

	// if the first token is an end anchor
	// it must match the end of the line
	if _, ok := regex[0].(EndAnchorClass); ok && len(regex) == 1 {
		return len(line) == 0 // end anchor matches only if line is empty
	}

	if len(line) == 0 {
		if regex[0].Optional() && len(regex) == 1 {
			return true
		}

		return false // no more characters in line to match against
	}

	if regex[0].AtLeastOne() {
		return matchPlus(regex[0], regex[1:], line)
	}

	if grp, ok := regex[0].(GroupClass); ok {
		for _, alt := range grp.alts {
			if matchHere(alt.tokens, line) {
				return true
			}
		}

		// If no alternatives matched, check if the group is optional
		// and try to match the rest of the regex
		if grp.Optional() {
			return matchHere(regex[1:], line)
		}

		return false // no alternative matched
	}

	if regex[0].Optional() {
		// If the next token is optional, we can either match it or skip it
		if regex[0].Check(line[0]) {
			if matchHere(regex[1:], line[1:]) {
				return true
			}
		}

		// Also try to match without the optional token
		return matchHere(regex[1:], line)
	}

	if regex[0].Check(line[0]) {
		return matchHere(regex[1:], line[1:])
	}

	return false
}

func matchPlus(c Class, regex []Class, line []byte) bool {
	// Match one or more occurrences of a class
	var i int
	for i < len(line) {
		// At least one class must match
		if grp, ok := c.(GroupClass); ok {
			found := false
			for _, alt := range grp.alts {
				if matchHere(alt.tokens, line) {
					found = true
					break
				}
			}

			if !found {
				return false // no alternative matched
			}
		} else if !c.Check(line[0]) {
			return false
		}

		if matchHere(regex, line[i+1:]) {
			return true
		}

		i++
	}

	return false
}

func matchLine(line []byte, pattern string) (bool, error) {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Fprintln(os.Stderr, "Logs from your program will appear here!")

	regex, err := Compile(pattern)
	if err != nil {
		return false, fmt.Errorf("error compiling regex: %v", err)
	}

	return match(regex, line), nil
}
