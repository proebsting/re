//  automata.go -- rx automata construction

//  Rx provides facilities for dealing with regular expressions.
package rx

import (
	"fmt"
)

var _ = fmt.Printf	//#%#%#% for debugging

// DFA is a deterministic finite automaton.
type DFA struct {
	//#%#% TBD
}

//  BuildDFA constructs a deterministic finite automaton from a parse tree.
//  In the process it modifies the parse tree, which is also returned.
func BuildDFA(tree Node) (*DFA, Node) {
	//#%#% augment the tree
	//#%#% split {m,n} nodes as necessary
	//#%#% number the leaves
	tree.SetNFL()    // set Nullable, FirstPos, LastPos values
	return nil, tree
}
