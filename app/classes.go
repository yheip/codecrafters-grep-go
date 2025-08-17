package main

import (
	"fmt"
	"slices"
)

type Class interface {
	Check(b byte) bool
	Optional() bool   // zero or one occurrence
	AtLeastOne() bool // one or more occurrences
	String() string
}

type CharClass struct {
	c byte
	Quantifier
}

func (c CharClass) Check(b byte) bool {
	return c.c == b
}

func (c CharClass) String() string {
	return string(c.c) + c.Quantifier.String()
}

type CharGroupClass struct {
	chars  []byte
	negate bool
	Quantifier
}

func (c CharGroupClass) Check(b byte) bool {
	if c.negate {
		return !slices.Contains(c.chars, b)
	}

	return slices.Contains(c.chars, b)
}

func (c CharGroupClass) String() string {
	opt := c.Quantifier.String()

	if c.negate {
		return fmt.Sprintf("[^%s]", string(c.chars)+opt)
	}

	return fmt.Sprintf("[%s]", string(c.chars)+opt)
}

type DigitClass struct {
	Quantifier
}

func (d DigitClass) Check(b byte) bool {
	return b >= '0' && b <= '9'
}

func (d DigitClass) String() string {
	return "\\d" + d.Quantifier.String()
}

type WordClass struct {
	Quantifier
}

func (w WordClass) Check(b byte) bool {
	return b == '_' || (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9')
}

func (w WordClass) String() string {
	return "\\w" + w.Quantifier.String()
}

type EndAnchorClass struct{}

func (e EndAnchorClass) Check(b byte) bool {
	return false
}

func (e EndAnchorClass) Optional() bool {
	return false
}
func (e EndAnchorClass) AtLeastOne() bool {
	return false
}

func (e EndAnchorClass) String() string {
	return "$"
}

// type PlusClass struct {
// 	OptionalClass
// }

// func (p PlusClass) Check(b byte) bool {
// 	return false
// }

// func (p PlusClass) String() string {
// 	if p.OptionalClass.Opt {
// 		return "+?"
// 	}

// 	return "+"
// }

type Quantifier struct {
	optional   bool
	atLeastOne bool
}

func (q Quantifier) Optional() bool {
	return q.optional
}

func (q Quantifier) AtLeastOne() bool {
	return q.atLeastOne
}

func (q Quantifier) String() string {
	str := ""

	if q.atLeastOne {
		str += "+"
	}

	if q.optional {
		str += "?"
	}

	return str
}

type OptionalClass struct {
	Opt bool
}

func (o OptionalClass) Optional() bool {
	return o.Opt
}

type WildcardClass struct {
	Quantifier
}

func (w WildcardClass) Check(b byte) bool {
	return true // Matches any character
}

func (w WildcardClass) String() string {
	str := "."

	if w.atLeastOne {
		str += "+"
	}

	if w.optional {
		str += "?"
	}

	return str
}

type GroupClass struct {
	alts []*Regex
	Quantifier
}

func (a GroupClass) Check(b byte) bool {
	return false // Alternation does not check individual bytes
}

func (a GroupClass) String() string {
	var result string
	for i, alt := range a.alts {
		if i > 0 {
			result += "|"
		}
		result += alt.String()
	}

	return "(" + result + ")" + a.Quantifier.String()
}

func printClass(c Class) {
	switch c := c.(type) {
	case CharClass:
		fmt.Printf("CharClass: %c, Optional: %t, AtLeastOne: %t\n", c.c, c.optional, c.atLeastOne)
	case DigitClass:
		fmt.Printf("DigitClass: Optional: %t, AtLeastOne: %t\n", c.optional, c.atLeastOne)
	case WordClass:
		fmt.Printf("WordClass: Optional: %t, AtLeastOne: %t\n", c.optional, c.atLeastOne)
	case CharGroupClass:
		fmt.Printf("CharGroupClass: chars: %s, negate: %t, Optional: %t, AtLeastOne: %t\n", string(c.chars), c.negate, c.optional, c.atLeastOne)
	case EndAnchorClass:
		fmt.Println("EndAnchorClass")
	case WildcardClass:
		fmt.Printf("WildcardClass: Optional: %t, AtLeastOne: %t\n", c.optional, c.atLeastOne)
	case GroupClass:
		fmt.Printf("GroupClass: %+v, Optional: %t, AtLeastOne: %t\n", c, c.optional, c.atLeastOne)
	default:
		fmt.Printf("Unknown token type: %T\n", c)
	}
}
