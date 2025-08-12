package main

import (
	"fmt"
	"slices"
)

type Class interface {
	Check(b byte) bool
	String() string
}

type CharClass struct {
	c byte
}

func (c CharClass) Check(b byte) bool {
	return c.c == b
}

func (c CharClass) String() string {
	return string(c.c)
}

type CharGroupClass struct {
	chars  []byte
	negate bool
}

func (c CharGroupClass) Check(b byte) bool {
	if c.negate {
		return !slices.Contains(c.chars, b)
	}

	return slices.Contains(c.chars, b)
}

func (c CharGroupClass) String() string {
	if c.negate {
		return fmt.Sprintf("[^%s]", string(c.chars))
	}

	return fmt.Sprintf("[%s]", string(c.chars))
}

type DigitClass struct{}

func (d DigitClass) Check(b byte) bool {
	return b >= '0' && b <= '9'
}

func (d DigitClass) String() string {
	return "\\d"
}

type WordClass struct{}

func (w WordClass) Check(b byte) bool {
	return b == '_' || (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9')
}

func (w WordClass) String() string {
	return "\\w"
}

type EndAnchorClass struct{}

func (e EndAnchorClass) Check(b byte) bool {
	return false
}

func (e EndAnchorClass) String() string {
	return "$"
}

type PlusClass struct{}

func (p PlusClass) Check(b byte) bool {
	return false
}

func (p PlusClass) String() string {
	return "+"
}
