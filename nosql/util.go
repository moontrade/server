package nosql

import (
	"sort"
	"strings"
)

// Copyright (c) 2017, A. Stoewer <adrian.stoewer@rz.ifi.lmu.de>
// All rights reserved.

// isLower checks if a character is lower case. More precisely it evaluates if it is
// in the range of ASCII character 'a' to 'z'.
func isLower(ch rune) bool {
	return ch >= 'a' && ch <= 'z'
}

// toLower converts a character in the range of ASCII characters 'A' to 'Z' to its lower
// case counterpart. Other characters remain the same.
func toLower(ch rune) rune {
	if ch >= 'A' && ch <= 'Z' {
		return ch + 32
	}
	return ch
}

// isUpper checks if a character is upper case. More precisely it evaluates if it is
// in the range of ASCII characters 'A' to 'Z'.
func isUpper(ch rune) bool {
	return ch >= 'A' && ch <= 'Z'
}

// isUpperChar checks if a character is upper case. More precisely it evaluates if it is
// in the range of ASCII characters 'A' to 'Z'.
func isUpperChar(ch byte) bool {
	return ch >= 'A' && ch <= 'Z'
}

// toLower converts a character in the range of ASCII characters 'a' to 'z' to its lower
// case counterpart. Other characters remain the same.
func toUpper(ch rune) rune {
	if ch >= 'a' && ch <= 'z' {
		return ch - 32
	}
	return ch
}

// isSpace checks if a character is some kind of whitespace.
func isSpace(ch rune) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}

// isDelimiter checks if a character is some kind of whitespace or '_' or '-'.
func isDelimiter(ch rune) bool {
	return ch == '-' || ch == '_' || isSpace(ch)
}

// iterFunc is a callback that is called fro a specific position in a string. Its arguments are the
// rune at the respective string position as well as the previous and the next rune. If curr is at the
// first position of the string prev is zero. If curr is at the end of the string next is zero.
type iterFunc func(prev, curr, next rune)

// stringIter iterates over a string, invoking the callback for every single rune in the string.
func stringIter(s string, callback iterFunc) {
	var prev rune
	var curr rune
	for _, next := range s {
		if curr == 0 {
			prev = curr
			curr = next
			continue
		}

		callback(prev, curr, next)

		prev = curr
		curr = next
	}

	if len(s) > 0 {
		callback(prev, curr, 0)
	}
}

// UpperCamelCase converts a string into camel case starting with a upper case letter.
func UpperCamelCase(s string) string {
	return camelCase(s, true)
}

// LowerCamelCase converts a string into camel case starting with a lower case letter.
func LowerCamelCase(s string) string {
	return camelCase(s, false)
}

func camelCase(s string, upper bool) string {
	s = strings.TrimSpace(s)
	buffer := make([]rune, 0, len(s))

	stringIter(s, func(prev, curr, next rune) {
		if !isDelimiter(curr) {
			if isDelimiter(prev) || (upper && prev == 0) {
				buffer = append(buffer, toUpper(curr))
			} else if isLower(prev) {
				buffer = append(buffer, curr)
			} else {
				buffer = append(buffer, toLower(curr))
			}
		}
	})

	return string(buffer)
}

// kebabCase converts a string into kebab case.
func kebabCase(s string) string {
	return delimiterCase(s, '-', false)
}

// upperKebabCase converts a string into kebab case with capital letters.
func upperKebabCase(s string) string {
	return delimiterCase(s, '-', true)
}

// snakeCase converts a string into snake case.
func snakeCase(s string) string {
	return delimiterCase(s, '_', false)
}

// upperSnakeCase converts a string into snake case with capital letters.
func upperSnakeCase(s string) string {
	return delimiterCase(s, '_', true)
}

// delimiterCase converts a string into snake_case or kebab-case depending on the delimiter passed
// as second argument. When upperCase is true the result will be UPPER_SNAKE_CASE or UPPER-KEBAB-CASE.
func delimiterCase(s string, delimiter rune, upperCase bool) string {
	s = strings.TrimSpace(s)
	buffer := make([]rune, 0, len(s)+3)

	adjustCase := toLower
	if upperCase {
		adjustCase = toUpper
	}

	var prev rune
	var curr rune
	for _, next := range s {
		if isDelimiter(curr) {
			if !isDelimiter(prev) {
				buffer = append(buffer, delimiter)
			}
		} else if isUpper(curr) {
			if isLower(prev) || (isUpper(prev) && isLower(next)) {
				buffer = append(buffer, delimiter)
			}
			buffer = append(buffer, adjustCase(curr))
		} else if curr != 0 {
			buffer = append(buffer, adjustCase(curr))
		}
		prev = curr
		curr = next
	}

	if len(s) > 0 {
		if isUpper(curr) && isLower(prev) && prev != 0 {
			buffer = append(buffer, delimiter)
		}
		buffer = append(buffer, adjustCase(curr))
	}

	return string(buffer)
}

// uint16Slice attaches the methods of Interface to []uint16, sorting in increasing order.
type uint16Slice []uint16

func (x uint16Slice) Len() int           { return len(x) }
func (x uint16Slice) Less(i, j int) bool { return x[i] < x[j] }
func (x uint16Slice) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }

// Sort is a convenience method: x.Sort() calls Sort(x).
func (x uint16Slice) Sort() { sort.Sort(x) }

// uint32Slice attaches the methods of Interface to []uint32, sorting in increasing order.
type uint32Slice []uint32

func (x uint32Slice) Len() int           { return len(x) }
func (x uint32Slice) Less(i, j int) bool { return x[i] < x[j] }
func (x uint32Slice) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }

// Sort is a convenience method: x.Sort() calls Sort(x).
func (x uint32Slice) Sort() { sort.Sort(x) }
