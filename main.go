package main

import (
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"
)

var capturedGroups map[int]string

func resetCapturedGroups() {
	capturedGroups = make(map[int]string)
}

// usage: echo <input_text> | grep_go.sh -E <pattern>
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
	patternLength := utf8.RuneCountInString(pattern)
	if patternLength == 0 {
		return false, fmt.Errorf("unsupported pattern: %q", pattern)
	}

	var ok bool

	if pattern[0] == byte('^') {
		resetCapturedGroups()
		return matchHere(line, pattern[1:]), nil
	}

	for i := range line {
		// sets the captured groups to be empty if moving onto the next character
		resetCapturedGroups()
		ok = matchHere(line[i:], pattern)
		if ok {
			return true, nil
		}
	}
	return false, nil
}

// make sure there is something in line that fits the pattern to capture then return the line index of where the last valid character is
func findValidCaptureGroup(line []byte, pattern string) int {
	// keep track of where in the line is valid
	linePos := 0
	for i, char := range pattern {
		// make sure the linePos does not go past how long the line is
		if linePos >= len(line) {
			return linePos
		}
		// if the previous character was \, skip this iteration (two character group like \w), and it is not the first check
		if i > 0 {
			prevChar, _ := utf8.DecodeRuneInString(pattern[i-1:])
			if prevChar == '\\' {
				continue
			}
		}
		switch char {
		// if the current character is +, this was handled in the last loop, so skip this one
		case '+':
			continue
		// if the current character is ?, this was handled in the last loop, so skip this one
		case '?':
			continue
		// determine what the pattern is and if it has a greedy character, and proceed accordingly
		case '\\':
			nextPatternChar, sizeNextPatternChar := utf8.DecodeRuneInString(pattern[i+1:])
			if nextPatternChar == utf8.RuneError {
				return -1
			}
			greedyCapture := false
			special, specialSize := utf8.DecodeRuneInString(pattern[i+sizeNextPatternChar+1:])
			if special != utf8.RuneError && special == '+' {
				greedyCapture = true
			}
			switch nextPatternChar {
			case 'w':
				switch greedyCapture {
				case true:
					valid := false
					for matchAlphaNumeric(rune(line[linePos]), "") {
						_, size := utf8.DecodeLastRuneInString(string(line))
						linePos += size
						valid = true
					}
					endOfPattern, _ := utf8.DecodeLastRuneInString(pattern[i+sizeNextPatternChar+specialSize:])
					if endOfPattern == utf8.RuneError {
						return linePos
					}
					if valid {
						continue
					}
					return -1

				case false:
					if matchAlphaNumeric(rune(line[linePos]), "") {
						_, size := utf8.DecodeLastRuneInString(string(line))
						linePos += size
						continue
					}
					if special != utf8.RuneError && special == '?' {
						continue
					}
					return -1
				}
			case 'd':
				switch greedyCapture {
				case true:
					valid := false
					for matchDigits(rune(line[linePos]), "") {
						_, size := utf8.DecodeLastRuneInString(string(line))
						linePos += size
						valid = true
					}
					endOfPattern, _ := utf8.DecodeLastRuneInString(pattern[i+sizeNextPatternChar+specialSize:])
					if endOfPattern == utf8.RuneError {
						return linePos
					}
					if valid {
						continue
					}
					return -1

				case false:
					if matchDigits(rune(line[linePos]), "") {
						_, size := utf8.DecodeLastRuneInString(string(line))
						linePos += size
						continue
					}
					if special != utf8.RuneError && special == '?' {
						continue
					}
					return -1
				}
			}
		// determine the kind of group and if it is greedy, and proceed accordingly
		case '[':
			endIdx := strings.IndexByte(pattern, ']')
			if endIdx == -1 {
				return -1
			}
			nextPatternChar, _ := utf8.DecodeRuneInString(pattern[i+1:])
			if nextPatternChar == utf8.RuneError {
				return -1
			}
			greedyCapture := false
			special, specialSize := utf8.DecodeRuneInString(string(pattern[endIdx+1]))
			if special != utf8.RuneError && special == '+' {
				greedyCapture = true
			}
			switch nextPatternChar {
			case '^':
				groupPat := pattern[2:endIdx]
				switch greedyCapture {
				case true:
					valid := false
					for matchNegativeGroup(rune(line[linePos]), groupPat) {
						_, size := utf8.DecodeLastRuneInString(string(line))
						linePos += size
						valid = true
					}
					endOfPattern, _ := utf8.DecodeLastRuneInString(pattern[endIdx+specialSize+1:])
					if endOfPattern == utf8.RuneError {
						return linePos
					}
					if valid {
						continue
					}
					return -1

				case false:
					if matchNegativeGroup(rune(line[linePos]), groupPat) {
						_, size := utf8.DecodeLastRuneInString(string(line))
						linePos += size
						continue
					}
					if special != utf8.RuneError && special == '?' {
						continue
					}
					return -1
				}

			default:
				groupPat := pattern[1:endIdx]
				switch greedyCapture {
				case true:
					valid := false
					for matchPositiveGroup(rune(line[linePos]), groupPat) {
						_, size := utf8.DecodeLastRuneInString(string(line))
						linePos += size
						valid = true
					}
					endOfPattern, _ := utf8.DecodeLastRuneInString(pattern[endIdx+specialSize+1:])
					if endOfPattern == utf8.RuneError {
						return linePos
					}
					if valid {
						continue
					}
					return -1

				case false:
					if matchPositiveGroup(rune(line[linePos]), groupPat) {
						_, size := utf8.DecodeLastRuneInString(string(line))
						linePos += size
						continue
					}
					if special != utf8.RuneError && special == '?' {
						continue
					}
					return -1
				}
			}
		// determine if it is a single or greedy wildcard, and proceed accordingly
		case '.':
			nextPatternChar, _ := utf8.DecodeRuneInString(pattern[i+1:])

			switch nextPatternChar {
			case '+':
				for linePos < len(line) {
					_, size := utf8.DecodeLastRuneInString(string(line))
					linePos += size
				}
				return linePos
			default:
				_, size := utf8.DecodeLastRuneInString(string(line))
				linePos += size
				continue
			}
		// if no special cases, check character against character and if it is greedy
		default:
			nextPatternChar, _ := utf8.DecodeRuneInString(pattern[i+1:])

			switch nextPatternChar {
			case '+':
				valid := false
				for linePos < len(line) && matchSameChar(rune(line[linePos]), string(char)) {
					_, size := utf8.DecodeLastRuneInString(string(line))
					linePos += size
					valid = true
				}
				if valid {
					continue
				}
				return -1
			case '?':
				if matchSameChar(rune(line[linePos]), string(char)) {
					_, size := utf8.DecodeLastRuneInString(string(line))
					linePos += size
					continue
				}
				continue
			default:
				if char == rune(line[linePos]) {
					_, size := utf8.DecodeLastRuneInString(string(line))
					linePos += size
					continue
				}
				return -1
			}
		}
	}
	return linePos
}

// recursive call to check if the pattern continues to be valid
func matchHere(line []byte, pattern string) bool {
	// If the pattern is empty, is has been successfully matched
	if len(pattern) == 0 {
		return true
	}

	// if the line is empty, check if the end of the pattern is '$' which signifies the end of that and false otherwise
	if len(line) == 0 {
		return pattern[0] == byte('$')
	}

	// get the char and its size from the first char in the line as well as how big the pattern char is
	char, size := utf8.DecodeRune(line)
	_, patternCharSize := utf8.DecodeRuneInString(pattern)
	currPatternLen := len(pattern)

	// check which pattern needs to be dealt with, and do so accordingly whether it is greedy or not
	// might be able to do away with len checks for something else
	switch {
	// digits pattern of \d
	case currPatternLen > 1 && pattern[0] == byte('\\') && pattern[1] == byte('d'):
		if currPatternLen > 2 && pattern[2] == byte('+') {
			return oneOrMore(line, "", pattern[3:], matchDigits)
		} else if matchDigits(char, "") {
			return matchHere(line[size:], pattern[2:])
		}
		return false
	// alphanumeric pattern of \w
	case currPatternLen > 1 && pattern[0] == byte('\\') && pattern[1] == byte('w'):
		if currPatternLen > 2 && pattern[2] == byte('+') {
			return oneOrMore(line, "", pattern[3:], matchAlphaNumeric)
		} else if matchAlphaNumeric(char, "") {
			return matchHere(line[size:], pattern[2:])
		}
		return false
	// wildcard pattern of .
	case pattern[0] == byte('.'):
		switch {
		case currPatternLen > 1 && pattern[1] == byte('+'):
			return true
		default:
			return matchHere(line[size:], pattern[1:])
		}

	// negative character group [^abc]
	case currPatternLen > 2 && pattern[0] == byte('[') && pattern[1] == byte('^'):
		end := strings.IndexByte(pattern, ']')
		negativeGroup := pattern[2:end]

		if end+1 < currPatternLen && pattern[end+1] == byte('+') {
			return oneOrMore(line, negativeGroup, pattern[end+2:], matchNegativeGroup)
		} else if !strings.ContainsRune(negativeGroup, rune(line[0])) {
			return matchHere(line[size:], pattern[end+1:])
		}
		return false

	// positive character group [abc]
	case currPatternLen > 1 && pattern[0] == byte('['):
		end := strings.IndexByte(pattern, ']')
		positiveGroup := pattern[1:end]
		if end+1 < currPatternLen && pattern[end+1] == byte('+') {
			return oneOrMore(line, positiveGroup, pattern[end+2:], matchPositiveGroup)
		} else if strings.ContainsRune(positiveGroup, rune(line[0])) {
			return matchHere(line[size:], pattern[end+1:])
		}
		return false

	// match one or more times pattern of + on a single character
	case currPatternLen > 1 && pattern[1] == byte('+'):
		return oneOrMore(line, string(pattern[0]), pattern[2:], matchSameChar)

	// match zero or one times pattern of ? on a single character
	case currPatternLen > 1 && pattern[1] == byte('?'):
		if line[0] == pattern[0] {
			return matchHere(line[size:], pattern[2:])
		}
		return matchHere(line, pattern[2:])

	// creates a new capture group with the pattern of (...) and checks for alternates with the pattern of (...'|'...)
	case currPatternLen > 2 && pattern[0] == byte('('):
		return handleCaptureGroup(line, pattern)

	// back-reference pattern of \1, \2, etc
	case currPatternLen > 1 && pattern[0] == byte('\\') && pattern[1] >= '1' && pattern[1] <= '9':
		groupToCheck := int(pattern[1] - '0')
		captured, ok := capturedGroups[groupToCheck]
		if ok && strings.HasPrefix(string(line), captured) {
			return matchHere(line[len(captured):], pattern[2:])
		}
		return false

	// no special pattern, check character against character
	default:
		if pattern[0] == line[0] {
			return matchHere(line[size:], pattern[patternCharSize:])
		}
	}
	// no valid matches, so this check fails
	return false
}

// determines the capture group and how to handle it
func handleCaptureGroup(line []byte, pattern string) bool {
	end := strings.IndexByte(pattern, ')')
	// if pattern is wrong, exit
	if end == -1 {
		return false
	}
	// gets the pattern in the ()
	groupPattern := pattern[1:end]
	splitPresent := strings.IndexByte(groupPattern, '|')
	if splitPresent != -1 {
		words := strings.Split(groupPattern, "|")
		for _, word := range words {
			if helperCaptureGroup(line, word, end, pattern) {
				return true
			}
		}
		return false
	}
	return helperCaptureGroup(line, groupPattern, end, pattern)
}

// tries to find a valid match for the capturn group pattern
func helperCaptureGroup(line []byte, groupPattern string, end int, pattern string) bool {
	// creates a new group with a distinct id
	newGroupId := len(capturedGroups) + 1
	// finds the idx of last valid line character
	idx := findValidCaptureGroup(line, groupPattern)
	// no match
	if idx == -1 {
		return false
	}
	// add capture group and continue recursion
	capturedGroups[newGroupId] = string(line[:idx])
	return matchHere(line[idx:], pattern[end+1:])
}

// helper functions to determine validity
func matchSameChar(char rune, charToMatch string) bool {
	return char == rune(charToMatch[0])
}

func matchDigits(char rune, _ string) bool {
	return strings.ContainsRune("0123456789", char)
}

func matchAlphaNumeric(char rune, _ string) bool {
	return strings.ContainsRune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_", char)
}

func matchPositiveGroup(char rune, group string) bool {
	return strings.ContainsRune(group, char)
}

func matchNegativeGroup(char rune, group string) bool {
	return !strings.ContainsRune(group, char)
}

// determines how many characters match the greedy character and continues the recursion from the proper index
func oneOrMore(line []byte, checkedChars string, pattern string, f func(rune, string) bool) bool {
	// There is nothing to check against
	if len(line) == 0 {
		return false
	}
	// keep track of valid sequence that matches the greedy character/pattern
	validSize := 0

	// keep checking validity as long as a character exists in line, and passes the proper check
	for {
		char, size := utf8.DecodeRune(line[validSize:])
		if char == utf8.RuneError {
			return matchHere(line[validSize:], pattern)
		} else if f(char, checkedChars) {
			validSize += size
			continue
		}
		if validSize == 0 {
			return false
		}
		return matchHere(line[validSize:], pattern)
	}
}
