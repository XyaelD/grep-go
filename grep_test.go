package main

import (
	"os/exec"
	"strings"
	"testing"
)

func runTest(t *testing.T, input string, pattern string, expectedExitCode int) {
	cmd := exec.Command("./grep_go.sh", "-E", pattern)
	cmd.Stdin = strings.NewReader(input)
	err := cmd.Run()
	exitCode := 0
	if err != nil {
		exitError, ok := err.(*exec.ExitError)
		if ok {
			exitCode = exitError.ExitCode()
		} else {
			t.Fatalf("Failed to run command: %v", err)
		}
	}
	if exitCode != expectedExitCode {
		t.Fatalf("Expected exit code %d, but got %d", expectedExitCode, exitCode)
	}
}

func TestGrepScript(t *testing.T) {
	tests := []struct {
		input            string
		pattern          string
		expectedExitCode int
	}{
		{"3 red squares and 3 red circles", "(\\d+) (\\w+) squares and \\1 \\2 circles", 0},
		{"3 red squares and 4 red circles", "(\\d+) (\\w+) squares and \\1 \\2 circles", 1},
		{"grep 101 is doing grep 101 times", "(\\w\\w\\w\\w) (\\d\\d\\d) is doing \\1 \\2 times", 0},
		{"$?! 101 is doing $?! 101 times", "(\\w\\w\\w) (\\d\\d\\d) is doing \\1 \\2 times", 1},
		{"grep yes is doing grep yes times", "(\\w\\w\\w\\w) (\\d\\d\\d) is doing \\1 \\2 times", 1},
		{"abc-def is abc-def, not efg", "([abc]+)-([def]+) is \\1-\\2, not [^xyz]+", 0},
		{"efg-hij is efg-hij, not efg", "([abc]+)-([def]+) is \\1-\\2, not [^xyz]+", 1},
		{"abc-def is abc-def, not xyz", "([abc]+)-([def]+) is \\1-\\2, not [^xyz]+", 1},
		{"apple pie, apple and pie", "^(\\w+) (\\w+), \\1 and \\2$", 0},
		{"pineapple pie, pineapple and pie", "^(apple) (\\w+), \\1 and \\2$", 1},
		{"apple pie, apple and pies", "^(\\w+) (pie), \\1 and \\2$", 1},
		{"howwdy hey there, howwdy hey", "(how+dy) (he?y) there, \\1 \\2", 0},
		{"hody hey there, howwdy hey", "(how+dy) (he?y) there, \\1 \\2", 1},
		{"howwdy heeey there, howwdy heeey", "(how+dy) (he?y) there, \\1 \\2", 1},
		{"cat and fish, cat with fish", "(c.t|d.g) and (f..h|b..d), \\1 with \\2", 0},
		{"bat and fish, cat with fish", "(c.t|d.g) and (f..h|b..d), \\1 with \\2", 1},
		{"cat and cat", "(cat) and \\1", 0},
		{"cat and dog", "(cat) and \\1", 1},
		{"grep 101 is doing grep 101 times", "(\\w\\w\\w\\w \\d\\d\\d) is doing \\1 times", 0},
		{"$?! 101 is doing $?! 101 times", "(\\w\\w\\w \\d\\d\\d) is doing \\1 times", 1},
		{"grep yes is doing grep yes times", "(\\w\\w\\w\\w \\d\\d\\d) is doing \\1 times", 1},
		{"abcd is abcd, not efg", "([abcd]+) is \\1, not [^xyz]+", 0},
		{"efgh is efgh, not efg", "([abcd]+) is \\1, not [^xyz]+", 1},
		{"abcd is abcd, not xyz", "([abcd]+) is \\1, not [^xyz]+", 1},
		{"this starts and ends with this", "^(\\w+) starts and ends with \\1$", 0},
		{"that starts and ends with this", "^(this) starts and ends with \\1$", 1},
		{"this starts and ends with this?", "^(this) starts and ends with \\1$", 1},
		{"once a dreaaamer, always a dreaaamer", "once a (drea+mer), alwaysz? a \\1", 0},
		{"once a dremer, always a dreaaamer", "once a (drea+mer), alwaysz? a \\1", 1},
		{"once a dreaaamer, alwayszzz a dreaaamer", "once a (drea+mer), alwaysz? a \\1", 1},
		{"bugs here and bugs there", "(b..s|c..e) here and \\1 there", 0},
		{"bugz here and bugs there", "(b..s|c..e) here and \\1 there", 1},
	}

	for _, test := range tests {
		t.Run(test.input+"_"+test.pattern, func(t *testing.T) {
			runTest(t, test.input, test.pattern, test.expectedExitCode)
		})
	}
}
