package main

import (
	"fmt"
	"io"
	"os"
	"slices"
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

type RE struct {
	classType classType
	chars     []byte
	negate    bool
}

type classType int

const (
	classTypeChar classType = iota
	classTypeCharGroup
	classTypeDigit
	classTypeWord
)

var (
	digitChars = []byte("0123456789")
	wordChars  = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_")
)

func Compile(pattern string) ([]RE, error) {
	var (
		i     int
		n     = len(pattern)
		regex = []RE{}
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
				regex = append(regex, RE{classType: classTypeDigit, chars: digitChars})
			case 'w':
				regex = append(regex, RE{classType: classTypeWord, chars: wordChars})
			default:
				// Treat as a literal character
				regex = append(regex, RE{
					classType: classTypeChar,
					chars:     []byte{pattern[i]},
				})
			}
		case '[':
			i++
			if i >= n {
				return nil, fmt.Errorf("unmatched '[' in pattern")
			}

			re := RE{
				classType: classTypeCharGroup,
			}

			if pattern[i] == '^' {
				i++
				re.negate = true
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
			re.chars = []byte(pattern[start:i])

			regex = append(regex, re)
		default:
			regex = append(regex, RE{
				classType: classTypeChar,
				chars:     []byte{pattern[i]},
			})
		}
		i++
	}

	return regex, nil
}

func match(regex []RE, line []byte) bool {
	var i int

	for i < len(line) {
		if matchHere(regex, line[i:]) {
			return true
		}

		i++
	}

	return false
}

func matchHere(regex []RE, line []byte) bool {
	if len(regex) == 0 {
		return true // empty regex matches everything
	}

	if len(line) == 0 {
		return false // no more characters in line to match against
	}

	switch regex[0].classType {
	case classTypeChar:
		if regex[0].chars[0] == line[0] {
			return matchHere(regex[1:], line[1:])
		}
	case classTypeDigit:
		if slices.Contains(regex[0].chars, line[0]) {
			return matchHere(regex[1:], line[1:])
		}
	case classTypeWord:
		if slices.Contains(regex[0].chars, line[0]) {
			return matchHere(regex[1:], line[1:])
		}
	case classTypeCharGroup:
		if regex[0].negate {
			if !slices.Contains(regex[0].chars, line[0]) {
				return matchHere(regex[1:], line[1:])
			}
		} else {
			if slices.Contains(regex[0].chars, line[0]) {
				return matchHere(regex[1:], line[1:])
			}
		}
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
