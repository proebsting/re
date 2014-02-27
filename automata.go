//  automata.go -- rx automata construction

//  Rx provides facilities for dealing with regular expressions.
package rx

import (
	"fmt"
)

var _ = fmt.Printf //#%#%#% for debugging

// DFA is a deterministic finite automaton.
type DFA struct {
	//#%#% TBD
}

//  BuildDFA constructs a deterministic finite automaton from a parse tree.
//  In the process it modifies the parse tree, which is also returned.
func BuildDFA(tree Node) (*DFA, Node) {

	// concatenate an Accept node to the end
	tree = Concatenate(tree, Accept())

	//#%#% split {m,n} nodes as necessary for a correct DFA

	// prepare nodes for followpos computation
	n := 0
	tree.Walk(nil, func(d Node) {
		d.SetNFL() // set Nullable, FirstPos, LastPos
		if leaf, ok := d.(*MatchNode); ok {
			n++ // number the leaf nodes
			leaf.posn = n
		}
	})

	// compute followpos sets
	tree.Walk(nil, func(d Node) {
		d.SetFollow()
	})

	//#%#% constuct and return DFA

	return nil, tree
}
