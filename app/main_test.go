package main

import (
	"testing"
)

type testcase struct {
	line     string
	pattern  string
	expected bool
}

func run(t *testing.T, line, pattern string, expected bool) {
	t.Run(line+"_"+pattern, func(t *testing.T) {
		result, err := matchLine([]byte(line), pattern)
		if err != nil {
			t.Errorf("Error matching line: %v", err)
		}
		if result != expected {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})
}

func Test_match_empty_pattern(t *testing.T) {
	tests := []testcase{
		{"", "", false},
		{"abc", "", true},
		{"123", "", true},
		{"123", "()", true},
	}

	for _, tt := range tests {
		run(t, tt.line, tt.pattern, tt.expected)
	}
}

func Test_matchLine(t *testing.T) {
	t.Run("match a literal character", func(t *testing.T) {
		tests := []testcase{
			{"abc", "a", true},
			{"123", "2", true},
			{"\\", "\\\\", true},
		}

		for _, tt := range tests {
			run(t, tt.line, tt.pattern, tt.expected)
		}
	})

	t.Run("match digits", func(t *testing.T) {
		tests := []testcase{
			{"123", "\\d", true},
			{"123", "\\d\\d", true},
			{"a", "\\d", false},
			{"", "\\w", false},
		}

		for _, tt := range tests {
			run(t, tt.line, tt.pattern, tt.expected)
		}
	})

	t.Run("match alphanumeric characters", func(t *testing.T) {
		tests := []testcase{
			{"a", "\\w", true},
			{"ab", "\\w\\w", true},
		}

		for _, tt := range tests {
			run(t, tt.line, tt.pattern, tt.expected)
		}
	})

	t.Run("positive character groups", func(t *testing.T) {
		tests := []testcase{
			{"dog", "[abc]", false},
			{"apple", "[abc]", true},
			{"bac", "[abc]", true},
		}

		for _, tt := range tests {
			run(t, tt.line, tt.pattern, tt.expected)
		}
	})

	t.Run("negative character groups", func(t *testing.T) {
		tests := []testcase{
			{"dog", "[^abc]", true},
			{"apple", "[^abc]", true},
			{"bac", "[^abc]", false},
		}

		for _, tt := range tests {
			run(t, tt.line, tt.pattern, tt.expected)
		}
	})
}

func Test_match_combining_multiple_classes(t *testing.T) {
	tests := []testcase{
		{"1 apple", "\\d apple", true},
		{"100 apples", "\\d\\d\\d apple", true},
		{"3 dogs", "\\d \\w\\w\\ws", true},
		{"4 cats", "\\d \\w\\w\\ws", true},
		{"1apples", "\\d\\d\\d apple", false},
		{"1 orange", "\\d apple", false},
		{"1 dog", "\\d \\w\\w\\ws", false},
	}

	for _, tt := range tests {
		run(t, tt.line, tt.pattern, tt.expected)
	}
}

func Test_match_start_of_string(t *testing.T) {
	tests := []testcase{
		{"log", "^log", true},
		{"slog", "^log", false},
	}

	for _, tt := range tests {
		run(t, tt.line, tt.pattern, tt.expected)
	}
}
func Test_match_end_of_string(t *testing.T) {
	tests := []testcase{
		{"dog", "dog$", true},
		{"dot", "do[tg]$", true},
		{"dog", "do[tg]$", true},
		{"dogs", "do[tg]$", false},
		{"dots", "do[tg]$", false},
		{"dogs", "dog$", false},
	}

	for _, tt := range tests {
		run(t, tt.line, tt.pattern, tt.expected)
	}
}

func Test_match_one_or_more(t *testing.T) {
	tests := []testcase{
		{"aats", "a+ts", true},
		{"caats", "ca+ts", true},
		{"caats", "c[a]+ts", true},
		{"caats", "c[a]+ts$", true},
		{"aat", "[a]+$t", false},
	}

	for _, tt := range tests {
		run(t, tt.line, tt.pattern, tt.expected)
	}
}

func Test_match_zero_or_one(t *testing.T) {
	tests := []testcase{
		{"dogs", "dogs?", true},
		{"dog", "dogs?", true},
		{"cat", "ca?t", true},
		{"act", "ca?t", true},
		{"cat", "dogs?", false},
	}

	for _, tt := range tests {
		run(t, tt.line, tt.pattern, tt.expected)
	}
}

func Test_match_wildcard(t *testing.T) {
	tests := []testcase{
		{"cat", "c.t", true},
		{"ct", "c.?t", true},
		{"cat", "c..t", false},
	}

	for _, tt := range tests {
		t.Run(tt.line+"_"+tt.pattern, func(t *testing.T) {
			result, err := matchLine([]byte(tt.line), tt.pattern)
			if err != nil {
				t.Errorf("Error matching line: %v", err)
			}
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func Test_match_alternation(t *testing.T) {
	tests := []testcase{
		{"cat", "(cat|dog)", true},
		{"cat", "(cat)s", false},
		{"cats", "(cat)+", true},
		{"dog", "(cat|dog)", true},
		{"bat", "(cat|dog)", false},
		{"dog", "(cat|dog)s", false},
		{"cat", "c(a|o)t", true},
		{"cot", "c(a|o)t", true},
		{"cut", "c(a|o)t", false},
		{"cat", "c([abc]|[123])t", true},
		{"c1t", "c([abc]|[123])t", true},
		{"catdog", "(cat|dog|cow)+", true},
		{"I see 1 cat", "^I see (\\d (cat|dog|cow)s?)", true},
		{"I see 1 cat, ", "^I see (\\d (cat|dog|cow)s?(, | and )?)", true},
		{"I see 1 cat and ", "^I see (\\d (cat|dog|cow)s?(, | and )?)", true},
		{"I see 1 cat, 2 dog", "^I see (\\d (cat|dog|cow)s?(, | and )?)+", true},
		{"I see 1 cat, 2 dogs and 3 cows", "^I see (\\d (cat|dog|cow)s?(, | and )?)+$", true},
		{"I see 1 cat, 2 dogs and 3 cows", "^I see (\\d (cat|dog|cow)(, | and )?)+$", false},
		{"cat dogs cows", "((cat|dog|cow)( )?)+$", false},
		{"ab", "(b)+$", true},
		{"ab", "(ab)+$", true},
	}

	for _, tt := range tests {
		t.Run(tt.line+"_"+tt.pattern, func(t *testing.T) {
			result, err := matchLine([]byte(tt.line), tt.pattern)
			if err != nil {
				t.Errorf("Error matching line: %v", err)
			}
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func Test_match_backreference(t *testing.T) {
	tests := []testcase{
		{"cat and cat", "(cat) and \\1", true},
		{"cat and dog", "(cat) and \\1", false},
		{"cat and cat", "(\\w+) and \\1", true},
		{"dog and dog", "(\\w+) and \\1", true},
		{"cat and dog", "(\\w+) and \\1", false},
	}

	for _, tt := range tests {
		run(t, tt.line, tt.pattern, tt.expected)
	}
}
