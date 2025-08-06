package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

// Ensures gofmt doesn't remove the "bytes" import above (feel free to remove this!)
var _ = bytes.ContainsAny

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

func matchLine(line []byte, pattern string) (bool, error) {
	var ok bool

	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Fprintln(os.Stderr, "Logs from your program will appear here!")

	var i int
	for i < len(pattern) {
		switch pattern[i] {
		case '\\':
			i++
			if i >= len(pattern) {
				return false, fmt.Errorf("incomplete escape sequence at end of pattern")
			}
			switch pattern[i] {
			case 'd':
				ok = bytes.ContainsAny(line, "0123456789")
			case 'w':
				ok = bytes.ContainsAny(line, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_")
			default:
				return false, fmt.Errorf("unknown escape sequence: \\%c", pattern[i])
			}
		case '[':
			i++
			start := i
			for i < len(pattern) && pattern[i] != ']' {
				i++
			}
			if i >= len(pattern) || pattern[i] != ']' {
				return false, fmt.Errorf("unmatched '[' in pattern")
			}
			end := i
			if start >= end {
				return false, fmt.Errorf("empty character group in pattern")
			}
			charGroup := pattern[start:end]
			ok = bytes.ContainsAny(line, charGroup)
		default:
			if !bytes.ContainsRune(line, rune(pattern[i])) {
				return false, nil // no match found
			} else {
				ok = true // at least one character matched
			}
		}
		i++
	}

	return ok, nil
}
