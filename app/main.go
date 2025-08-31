package main

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/codecrafters-io/grep-starter-go/app/matcher"
	"github.com/codecrafters-io/grep-starter-go/app/parser"
	"github.com/codecrafters-io/grep-starter-go/app/regex"
)

// Usage: echo <input_text> | your_program.sh -E <pattern>
func main() {
	if len(os.Args) < 3 || os.Args[1] != "-E" {
		fmt.Fprintf(os.Stderr, "usage: mygrep -E <pattern>\n")
		os.Exit(2) // 1 means no lines were selected, >1 means error
	}

	pattern := os.Args[2]

	// If files are provided, search within those files line-by-line.
	if len(os.Args) > 3 {
		files := os.Args[3:]
		multi := len(files) > 1
		foundAny := false

		for _, fname := range files {
			f, err := os.Open(fname)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: open file %s: %v\n", fname, err)
				os.Exit(2)
			}

			scanner := bufio.NewScanner(f)
			// Increase the buffer limit to handle long lines (up to 10MB)
			buf := make([]byte, 0, 64*1024)
			scanner.Buffer(buf, 10*1024*1024)

			for scanner.Scan() {
				text := scanner.Text()
				ok, err := matchLine([]byte(text), pattern)
				if err != nil {
					_ = f.Close()
					fmt.Fprintf(os.Stderr, "error: %v\n", err)
					os.Exit(2)
				}
				if ok {
					foundAny = true
					if multi {
						fmt.Printf("%s:%s\n", fname, text)
					} else {
						fmt.Println(text)
					}
				}
			}

			if err := scanner.Err(); err != nil {
				_ = f.Close()
				fmt.Fprintf(os.Stderr, "error: scan file %s: %v\n", fname, err)
				os.Exit(2)
			}
			_ = f.Close()
		}

		if !foundAny {
			os.Exit(1)
		}
		return // success
	}

	// Fallback: read from stdin as a single line (legacy behavior).
	line, err := io.ReadAll(os.Stdin)
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
	if len(line) == 0 && pattern == "" {
		return false, nil
	}

	p := parser.New(pattern)
	regexNode, err := p.Parse()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: parse pattern: %v\n", err)
		os.Exit(2)
	}
	re, err := regex.Compile(regexNode)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: compile pattern: %v\n", err)
		os.Exit(2)
	}

	if re == nil {
		return false, fmt.Errorf("compiled regex is nil")
	}

	return matcher.Match(line, re), nil
}
