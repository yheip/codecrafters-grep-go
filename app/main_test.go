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
		{"dogs", "dog$", false},
	}

	for _, tt := range tests {
		run(t, tt.line, tt.pattern, tt.expected)
	}
}

func Test_match_one_or_more(t *testing.T) {
	tests := []testcase{
		{"caats", "ca+ts", true},
		{"caats", "c[a]+ts", true},
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
		{"dog", "(cat|dog)", true},
		{"bat", "(cat|dog)", false},
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
		{"I see 1 cat, 2 dogs and 3 cows", "^I see (\\d (cat|dog|cow)s?(, | and )?)+$", true},
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

func Test_Compile(t *testing.T) {
	tests := []struct {
		pattern string
		result  *Regex
	}{
		{"", &Regex{tokens: []Class{}}},
		{"\\d", &Regex{tokens: []Class{DigitClass{}}}},
		{"\\w", &Regex{tokens: []Class{WordClass{}}}},
		{"\\w\\d", &Regex{tokens: []Class{WordClass{}, DigitClass{}}}},
		{"[abc]", &Regex{tokens: []Class{CharGroupClass{chars: []byte{'a', 'b', 'c'}}}}},
		{"[^abc]", &Regex{tokens: []Class{CharGroupClass{negate: true, chars: []byte{'a', 'b', 'c'}}}}},
		{"^abc", &Regex{tokens: []Class{CharClass{c: 'a'}, CharClass{c: 'b'}, CharClass{c: 'c'}}, matchStart: true}},
		{"$", &Regex{tokens: []Class{EndAnchorClass{}}}},
		{".", &Regex{tokens: []Class{WildcardClass{}}}},
		{"a+", &Regex{tokens: []Class{CharClass{'a', Quantifier{atLeastOne: true}}}}},
		{"a+?", &Regex{tokens: []Class{CharClass{'a', Quantifier{optional: true, atLeastOne: true}}}}},
		{"(cat|dog)", &Regex{tokens: []Class{
			GroupClass{
				alts: []*Regex{
					{tokens: []Class{CharClass{c: 'c'}, CharClass{c: 'a'}, CharClass{c: 't'}}},
					{tokens: []Class{CharClass{c: 'd'}, CharClass{c: 'o'}, CharClass{c: 'g'}}},
				},
			},
		}}},
		{"my(cat|dog)", &Regex{tokens: []Class{
			CharClass{c: 'm'},
			CharClass{c: 'y'},
			GroupClass{
				alts: []*Regex{
					{tokens: []Class{CharClass{c: 'c'}, CharClass{c: 'a'}, CharClass{c: 't'}}},
					{tokens: []Class{CharClass{c: 'd'}, CharClass{c: 'o'}, CharClass{c: 'g'}}},
				},
			},
		}}},
		{"(cat|dog|bat)s", &Regex{tokens: []Class{
			GroupClass{
				alts: []*Regex{
					{tokens: []Class{CharClass{c: 'c'}, CharClass{c: 'a'}, CharClass{c: 't'}}},
					{tokens: []Class{CharClass{c: 'd'}, CharClass{c: 'o'}, CharClass{c: 'g'}}},
					{tokens: []Class{CharClass{c: 'b'}, CharClass{c: 'a'}, CharClass{c: 't'}}},
				},
			},
			CharClass{c: 's'},
		}}},
		{"(cat|[dog])", &Regex{tokens: []Class{
			GroupClass{
				alts: []*Regex{
					{tokens: []Class{CharClass{c: 'c'}, CharClass{c: 'a'}, CharClass{c: 't'}}},
					{tokens: []Class{CharGroupClass{chars: []byte{'d', 'o', 'g'}}}},
				},
			},
		}}},
		{"(cat)", &Regex{tokens: []Class{
			GroupClass{
				alts: []*Regex{
					{tokens: []Class{CharClass{c: 'c'}, CharClass{c: 'a'}, CharClass{c: 't'}}},
				},
			},
		}}},
		{"(cat)?", &Regex{tokens: []Class{
			GroupClass{
				alts: []*Regex{
					{tokens: []Class{CharClass{c: 'c'}, CharClass{c: 'a'}, CharClass{c: 't'}}},
				},
				Quantifier: Quantifier{optional: true},
			},
		}}},
		{"cat(dog)", &Regex{tokens: []Class{
			CharClass{c: 'c'}, CharClass{c: 'a'}, CharClass{c: 't'},
			GroupClass{
				alts: []*Regex{
					{tokens: []Class{CharClass{c: 'd'}, CharClass{c: 'o'}, CharClass{c: 'g'}}},
				},
			},
		}}},
		{"(cat)dog", &Regex{tokens: []Class{
			GroupClass{
				alts: []*Regex{
					{tokens: []Class{CharClass{c: 'c'}, CharClass{c: 'a'}, CharClass{c: 't'}}},
				},
			},
			CharClass{c: 'd'}, CharClass{c: 'o'}, CharClass{c: 'g'},
		}}},
		{"(cat(dog)(cow))", &Regex{tokens: []Class{
			GroupClass{
				alts: []*Regex{
					{tokens: []Class{
						CharClass{c: 'c'}, CharClass{c: 'a'}, CharClass{c: 't'},
						GroupClass{
							alts: []*Regex{
								{tokens: []Class{CharClass{c: 'd'}, CharClass{c: 'o'}, CharClass{c: 'g'}}},
							},
						},
						GroupClass{
							alts: []*Regex{
								{tokens: []Class{CharClass{c: 'c'}, CharClass{c: 'o'}, CharClass{c: 'w'}}},
							},
						},
					}},
				},
			},
		}}},
		{"(\\d (cat|dog|cow)s?(, | and )?)", &Regex{tokens: []Class{
			GroupClass{
				alts: []*Regex{
					{tokens: []Class{DigitClass{}, CharClass{c: ' '},
						GroupClass{
							alts: []*Regex{
								{tokens: []Class{CharClass{c: 'c'}, CharClass{c: 'a'}, CharClass{c: 't'}}},
								{tokens: []Class{CharClass{c: 'd'}, CharClass{c: 'o'}, CharClass{c: 'g'}}},
								{tokens: []Class{CharClass{c: 'c'}, CharClass{c: 'o'}, CharClass{c: 'w'}}},
							},
						},
						CharClass{
							c:          's',
							Quantifier: Quantifier{true, false},
						},
						GroupClass{
							alts: []*Regex{
								{tokens: []Class{CharClass{c: ','}, CharClass{c: ' '}}},
								{tokens: []Class{CharClass{c: ' '}, CharClass{c: 'a'}, CharClass{c: 'n'}, CharClass{c: 'd'}, CharClass{c: ' '}}},
							},
							Quantifier: Quantifier{true, false},
						},
					}},
				},
			},
		}}},
	}

	for _, tt := range tests {
		t.Run(tt.pattern, func(t *testing.T) {
			result, err := Compile(tt.pattern)
			if err != nil {
				t.Errorf("Error compiling pattern: %v", err)
			}
			if len(result.tokens) != len(tt.result.tokens) {
				t.Errorf("Expected %d tokens, got %d", len(tt.result.tokens), len(result.tokens))
			}

			if result.String() != tt.result.String() {
				t.Errorf("Expected %s, got %s", tt.result.String(), result.String())
			}
		})
	}
}
