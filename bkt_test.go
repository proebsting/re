//  bkt_test.go -- unit tests for bracket-expression parsing
//
//  These tests validate bxparse() and also BitSet.Bracketed().
//
//  These tests are run by the "go test" command.

package rx

import (
	"fmt"
	"testing"
)

//  TestBrackets runs a series of self-contained small tests.
func TestBrackets(t *testing.T) {
	fmt.Println("bkt_test.go: TestBrackets")
	// try(t, string, expected)
	try(t, `[a]`, `[a]`)
	try(t, `[bc]`, `[bc]`)
	try(t, `[def]`, `[def]`)
	try(t, `[ghij]`, `[g-j]`)
	try(t, `[lmnop]`, `[l-p]`)
	try(t, `[tuvwxyz]`, `[t-z]`)
	try(t, `[ACDFGHJKLMOPQRSUVWXYZ]`, `[ACDFGHJ-MO-SU-Z]`)
	try(t, `[aeiuo]`, `[aeiou]`)
	try(t, `[aeiuo]`, `[aeiou]`)
	try(t, `[a-z]`, `[a-z]`)
	try(t, `[A-Z]`, `[A-Z]`)
	try(t, `[a-zA-Z]`, `[A-Za-z]`)
	try(t, `[_a-zA-Z]`, `[A-Z_a-z]`)
	try(t, `[0-9]`, `[0-9]`)
	try(t, `[_a-zA-Z0-9]`, `[0-9A-Z_a-z]`)
	try(t, `[_a-zA-Z]`, `[A-Z_a-z]`)
	try(t, `[A-HO-Z]`, `[A-HO-Z]`)
	// ugly but legal cset (bracket expression) forms
	try(t, `[-]`, `[-]`)
	try(t, `[-x]`, `[-x]`)
	try(t, `[x-]`, `[-x]`)
	try(t, `[[x]`, `[[x]`)
	try(t, `[[x-]`, `[-[x]`)
	try(t, `[]x]`, `[]x]`)
	try(t, `[]x-]`, `[-]x]`)
	try(t, `[]x[-]`, `[-[]x]`)
	try(t, `[][]`, `[[]]`)
	// character set escapes
	// n.b. in POSIX, \ in [] is not special, but we follow Perl here
	try(t, `[\-]`, `[-]`)
	try(t, `[\]]`, `[]]`)
	try(t, `[ab\[cd\-gh\]ij]`, `[-[]a-dg-j]`)
	try(t, `[*&=!+]`, `[!&*+=]`)
	try(t, `[\*\&\=\!\+]`, `[!&*+=]`)
	// perl inventions
	try(t, `[\d]`, `[0-9]`)
	try(t, `[\d0IZESB]`, `[0-9BEISZ]`)
	try(t, `[\w]`, `[0-9A-Z_a-z]`)
	// C-style escapes
	try(t, `[\a]`, `[\a]`)
	try(t, `[\b]`, `[\b]`)
	try(t, `[\e]`, `[\x1b]`)
	try(t, `[\f]`, `[\f]`)
	try(t, `[\n]`, `[\n]`)
	try(t, `[\r]`, `[\r]`)
	try(t, `[\t]`, `[\t]`)
	try(t, `[\v]`, `[\v]`)
	try(t, `[\y]`, `'\y' unrecognized`)
	// octal escapes
	try(t, `[\0]`, `[\x00]`)
	try(t, `[\00]`, `[\x00]`)
	try(t, `[\000]`, `[\x00]`)
	try(t, `[\1a]`, `[\x01a]`)
	try(t, `[\11b]`, `[\tb]`)
	try(t, `[\111c]`, `[Ic]`)
	try(t, `[\7]`, `[\a]`)
	try(t, `[\043]`, `[#]`)
	try(t, `[\45]`, `[%]`)
	try(t, `[\75]`, `[=]`)
	try(t, `[\107]`, `[G]`)
	try(t, `[\176]`, `[~]`)
	// hex escapes
	try(t, `[\x2a3]`, `[*3]`)
	try(t, `[\u006B4]`, `[4k]`)
	try(t, `[\x`, `malformed '\xhh'`)
	try(t, `[\x5`, `malformed '\xhh'`)
	try(t, `[\x5G`, `malformed '\xhh'`)
	try(t, `[\u`, `malformed '\uhhhh'`)
	try(t, `[\u173]`, `malformed '\uhhhh'`)
	try(t, `[\umass]`, `malformed '\uhhhh'`)
	// big sets \D \W \S
	try(t, `[\D]`, `[\x01-/:-\u007f]`)
	try(t, `[\W]`, "[\\x01-/:-@[-^`{-\\u007f]")
	try(t, `[\S]`, `[\x01-\b\x0e-\x1f!-\u007f]`)
	// the following pairs should be identical
	try(t, `[\d]`, `[0-9]`)
	try(t, `[^\D]`, `[0-9]`)
	try(t, `[\w]`, `[0-9A-Z_a-z]`)
	try(t, `[^\W]`, `[0-9A-Z_a-z]`)
	try(t, `[\s]`, `[\t-\r ]`)
	try(t, `[^\S]`, `[\t-\r ]`)
	// complemented character sets
	try(t, "[^ -`]", `[\x01-\x1fa-\u007f]`)
	try(t, `[^ -@]`, `[\x01-\x1fA-\u007f]`)
	try(t, `[^ -/]`, `[\x01-\x1f0-\u007f]`)
	try(t, "[^ -/:-@[-`{-~]", `[\x01-\x1f0-9A-Za-z\u007f]`)
	try(t, `[^0-9]`, `[\x01-/:-\u007f]`)
	try(t, `[^A-Za-z0-9]`, "[\\x01-/:-@[-`{-\\u007f]")
	// these last two should be identical (under -Z)
	try(t, `[^A-Za-z0-9_]`, "[\\x01-/:-@[-^`{-\\u007f]")
	try(t, `[^\w]`, "[\\x01-/:-@[-^`{-\\u007f]")
	// erroneous
	try(t, `[`, `unclosed '['`)
	try(t, `[^`, `unclosed '['`)
	try(t, `[]`, `unclosed '['`)
	try(t, `[^]`, `unclosed '['`)
	try(t, `[\]`, `unclosed '['`)
	try(t, `[abc`, `unclosed '['`)
	try(t, `[def\]`, `unclosed '['`)
}

//  try parses one string and checks the result.
//  The input string should begin with the initial '[' to be ignored.
func try(t *testing.T, input string, expected string) {
	var output string
	subj := input[1:]
	result, errmsg := bxparse(subj)
	if result == nil {
		output = errmsg
	} else {
		output = result.Bracketed()
	}
	if output != expected {
		t.Error(input, "=>", output, "(expected ", expected, ")")
	}
}
