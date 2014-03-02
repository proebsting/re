//  rx.go -- top-level entry points

//  Rx provides facilities for dealing with regular expressions.
package rx

import (
)

//  Match tests whether a string is matched by a regular expression.
func Match(rexpr string, s string) (bool, error) {
	tree, err := Parse(rexpr)
	if err != nil {
		return false, err
	}
	dfa, _ := BuildDFA(tree)
	return dfa.Accepts(s), nil
}
