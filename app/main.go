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

func Compile(pattern string) (*Regex, error) {
	var (
		i     int
		n     = len(pattern)
		regex = &Regex{
			tokens: []Class{},
		}
	)

	for i < n {
		switch pattern[i] {
		case '\\':
			i++
			if i >= n {
				return nil, fmt.Errorf("incomplete escape sequence at end of pattern")
			}
			switch pattern[i] {
			case 'd':
				regex.tokens = append(regex.tokens, DigitClass{})
			case 'w':
				regex.tokens = append(regex.tokens, WordClass{})
			default:
				// Treat as a literal character
				regex.tokens = append(regex.tokens, CharClass{c: pattern[i]})
			}
		case '[':
			i++
			if i >= n {
				return nil, fmt.Errorf("unmatched '[' in pattern")
			}

			c := CharGroupClass{}

			if pattern[i] == '^' {
				i++
				c.negate = true
			}

			start := i
			for i < n && pattern[i] != ']' {
				i++
			}

			if i >= n || pattern[i] != ']' {
				return nil, fmt.Errorf("unmatched '[' in pattern")
			}

			if start >= i {
				return nil, fmt.Errorf("empty character group in pattern")
			}
			c.chars = []byte(pattern[start:i])

			regex.tokens = append(regex.tokens, c)
		case '^':
			regex.matchStart = true
			if i > 0 {
				return nil, fmt.Errorf("caret '^' must be at the start of the pattern")
			}
		case '$':
			regex.tokens = append(regex.tokens, EndAnchorClass{})
			if i < n-1 {
				return nil, fmt.Errorf("dollar '$' must be at the end of the pattern")
			}
		case '+':
			regex.tokens = append(regex.tokens, PlusClass{})
		default:
			regex.tokens = append(regex.tokens, CharClass{c: pattern[i]})
		}
		i++
	}

	return regex, nil
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
		return false // no more characters in line to match against
	}

	if len(regex) > 1 {
		if _, ok := regex[1].(PlusClass); ok {
			// If the next token is a plus, we need to match one or more occurrences
			return matchPlus(regex[0], regex[2:], line)
		}
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
		if !c.Check(line[i]) {
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
