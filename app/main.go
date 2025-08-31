package main

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/codecrafters-io/grep-starter-go/app/matcher"
	"github.com/codecrafters-io/grep-starter-go/app/parser"
	"github.com/codecrafters-io/grep-starter-go/app/regex"
)

// Usage: echo <input_text> | your_program.sh -E <pattern>
func main() {
	// Parse flags: support -E <pattern> [paths...] and optional -r for recursive directory search.
	recursive, pattern, paths, err := parseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(2)
	}

	// Compile regex once.
	re, err := compilePattern(pattern)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(2)
	}

	// If -r is set, we expect at least one path (directory or file) and always print with filename prefix.
	if recursive {
		if len(paths) == 0 {
			fmt.Fprintf(os.Stderr, "usage: mygrep -r -E <pattern> <path> [<path> ...]\n")
			os.Exit(2)
		}
		foundAny := false
		for _, p := range paths {
			info, statErr := os.Stat(p)
			if statErr != nil {
				fmt.Fprintf(os.Stderr, "error: stat path %s: %v\n", p, statErr)
				os.Exit(2)
			}
			if info.IsDir() {
				walkErr := filepath.WalkDir(p, func(path string, d fs.DirEntry, walkErr error) error {
					if walkErr != nil {
						return walkErr
					}
					if d.IsDir() {
						return nil
					}
					// Only process regular files
					if !d.Type().IsRegular() {
						return nil
					}
					matched, procErr := processFile(path, re, true)
					if procErr != nil {
						return procErr
					}
					if matched {
						foundAny = true
					}
					return nil
				})
				if walkErr != nil {
					fmt.Fprintf(os.Stderr, "error: walk path %s: %v\n", p, walkErr)
					os.Exit(2)
				}
			} else {
				matched, procErr := processFile(p, re, true)
				if procErr != nil {
					fmt.Fprintf(os.Stderr, "error: process file %s: %v\n", p, procErr)
					os.Exit(2)
				}
				if matched {
					foundAny = true
				}
			}
		}
		if !foundAny {
			os.Exit(1)
		}
		return
	}

	if len(paths) > 0 {
		multi := len(paths) > 1
		foundAny := false
		for _, fname := range paths {
			matched, err := processFile(fname, re, multi)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: process file %s: %v\n", fname, err)
				os.Exit(2)
			}
			if matched {
				foundAny = true
			}
		}
		if !foundAny {
			os.Exit(1)
		}
		return
	}

	line, rerr := io.ReadAll(os.Stdin)
	if rerr != nil {
		fmt.Fprintf(os.Stderr, "error: read input text: %v\n", rerr)
		os.Exit(2)
	}
	if !matcher.Match(line, re) {
		os.Exit(1)
	}
}

func matchWithCompiled(line []byte, re *regex.CompiledRegex) bool {
	if re == nil {
		return false
	}
	return matcher.Match(line, re)
}

// matchLine keeps the original test-facing API: compile pattern, then match once.
func matchLine(line []byte, pattern string) (bool, error) {
	if len(line) == 0 && pattern == "" {
		return false, nil
	}
	re, err := compilePattern(pattern)
	if err != nil {
		return false, err
	}
	return matchWithCompiled(line, re), nil
}

// compilePattern parses and compiles a pattern once.
func compilePattern(pattern string) (*regex.CompiledRegex, error) {
	p := parser.New(pattern)
	regexNode, err := p.Parse()
	if err != nil {
		return nil, fmt.Errorf("parse pattern: %w", err)
	}
	re, err := regex.Compile(regexNode)
	if err != nil {
		return nil, fmt.Errorf("compile pattern: %w", err)
	}
	if re == nil {
		return nil, fmt.Errorf("compiled regex is nil")
	}
	return re, nil
}

// processFile scans a file line-by-line and prints matches. If alwaysPrefix is true, prefix filename for each matched line.
// Returns whether any match was found in this file.
func processFile(path string, re *regex.CompiledRegex, alwaysPrefix bool) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	// Increase the buffer limit to handle long lines (up to 10MB)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 10*1024*1024)

	found := false
	for scanner.Scan() {
		text := scanner.Text()
		if matchWithCompiled([]byte(text), re) {
			found = true
			if alwaysPrefix {
				fmt.Printf("%s:%s\n", path, text)
			} else {
				fmt.Println(text)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return found, fmt.Errorf("scan file: %w", err)
	}
	return found, nil
}

// parseArgs parses supported CLI flags and returns (recursive, pattern, paths, error).
func parseArgs(args []string) (bool, string, []string, error) {
	recursive := false
	pattern := ""
	paths := []string{}

	i := 0
	for i < len(args) {
		a := args[i]
		switch a {
		case "-r":
			recursive = true
			i++
		case "-E":
			if i+1 >= len(args) {
				return false, "", nil, fmt.Errorf("usage: mygrep [-r] -E <pattern> [<path> ...]")
			}
			pattern = args[i+1]
			i += 2
		default:
			paths = append(paths, a)
			i++
		}
	}

	if pattern == "" {
		return false, "", nil, fmt.Errorf("usage: mygrep [-r] -E <pattern> [<path> ...]")
	}
	return recursive, pattern, paths, nil
}
