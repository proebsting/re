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

	// concatenate an Accept node to the end, without collapsing
	nlist := append(make([]Node, 0, 2), tree, Accept())
	tree = &ConcatNode{nlist, nildata}

	//#%#% split {m,n} nodes as necessary for a correct DFA

	// number the leaf nodes for better comprehension of the DFA
	n := 0
	tree.Walk(nil, nil, func(d Node, v interface{}) {
		if leaf, ok := d.(*MatchNode); ok {
			n++	
			leaf.posn = n 
		}
	})

	tree.SetNFL() // set Nullable, FirstPos, LastPos values

	return nil, tree
}
