//  rx.go -- top-level entry points

//  Rx provides facilities for dealing with regular expressions.
package rx

import ()

//  Match tests whether a string is matched by a regular expression.
func Match(rexpr string, s string) (bool, error) {
	dfa, err := Compile(rexpr)
	if err != nil {
		return false, err
	}
	return dfa.Accepts(s), nil
}

//  Compile makes a DFA from a regular expression.
//  The DFA can be reused for multiple matches, saving compilation costs.
func Compile(rexpr string) (*DFA, error) {
	ptree, err := Parse(rexpr) // make faithful parse tree
	if err != nil {
		return nil, err
	}
	atree := Augment(ptree) // make augmented parse tree
	dfa := BuildDFA(atree)  // build DFA from augmented tree
	return dfa, nil

}

//  Augment produces a modified parse tree in preparation for building a DFA.
//  The original tree is modified in place, and a new root is returned.
func Augment(tree Node) Node {
	tree = Concatenate(tree, Accept())
	//#%#% more work is needed to handle fixed-count replication
	return tree
}
