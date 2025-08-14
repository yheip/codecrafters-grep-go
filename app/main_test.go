package main

import "testing"

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

func Test_matchLine(t *testing.T) {
	t.Run("match a literal character", func(t *testing.T) {
		tests := []testcase{
			{"abc", "", true},
			{"123", "", true},
			{"\\", "\\\\", true},
			{"", "", false},
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

	t.Run("combining multiple character classes", func(t *testing.T) {
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
	})

	t.Run("match start of string anchor", func(t *testing.T) {
		tests := []testcase{
			{"log", "^log", true},
			{"slog", "^log", false},
		}

		for _, tt := range tests {
			run(t, tt.line, tt.pattern, tt.expected)
		}
	})

	t.Run("match end of string anchor", func(t *testing.T) {
		tests := []testcase{
			{"dog", "dog$", true},
			{"dogs", "dog$", false},
		}

		for _, tt := range tests {
			run(t, tt.line, tt.pattern, tt.expected)
		}
	})

	t.Run("Match one or more times", func(t *testing.T) {
		tests := []testcase{
			{"caats", "ca+ts", true},
			{"caats", "c[a]+ts", true},
		}

		for _, tt := range tests {
			run(t, tt.line, tt.pattern, tt.expected)
		}
	})

	t.Run("Match zero or one times", func(t *testing.T) {
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
	})
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
