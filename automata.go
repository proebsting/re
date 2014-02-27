//  automata.go -- rx automata construction

//  Rx provides facilities for dealing with regular expressions.
package rx

import (
	"fmt"
)

var _ = fmt.Printf //#%#%#% for debugging

// DFA is a deterministic finite automaton.
type DFA struct {
	Leaves []*MatchNode	// leaves (positions) from parse tree
	Dstates[]*DFAstate	// sets of posiitons
}

// DFAstate is one state in a DFA
type DFAstate struct {
	Posns map[*MatchNode]bool	// set of positions in the state
	Dnext map[byte]DFAstate		// transition map
}


//  BuildDFA constructs a deterministic finite automaton from a parse tree.
//  In the process it modifies the parse tree, which is also returned.
func BuildDFA(tree Node) (*DFA, Node) {

	dfa := &DFA{make([]*MatchNode, 0), make([]*DFAstate, 0)}

	// concatenate an Accept node to the end
	tree = Concatenate(tree, Accept())

	//#%#% Need to split {m,n} nodes as necessary for a correct DFA.
	//#%#% I *think* need to split if bounded maximum and/or minimum > 1.

	// prepare nodes for followpos computation
	n := 0
	tree.Walk(nil, func(d Node) {
		d.SetNFL() // set Nullable, FirstPos, LastPos
		if leaf, ok := d.(*MatchNode); ok {
			n++ // number the leaf nodes
			leaf.Posn = n
			dfa.Leaves = append(dfa.Leaves, leaf)
		}
	})

	// compute followpos sets
	tree.Walk(nil, func(d Node) {
		d.SetFollow()
	})

	// compute DFA; see Dragon2 p. 141

	//#%#%#% set first state

	for nmarked := 0; nmarked < len(dfa.Dstates); nmarked++ {
		d := dfa.Dstates[nmarked]
		alist := followchars(d)
		for a := range alist {
			u := followposns(d, a)
			u = u	//#%#%#%
		}
	}

	// return DFA and augmented tree
	return dfa, tree
}

func followchars(d *DFAstate) string {
	cs := CharSet("")
	for p := range d.Posns {
		for q := range p.FollowPos {
			cs = cs.Or(q.cset)
		}
	}
	return cs.Members()
}

func followposns(d *DFAstate, a int) map[*MatchNode]bool {
	posns := make(map[*MatchNode]bool)
	for p := range d.Posns {
		for q := range p.FollowPos {
			if q.cset.Test(uint(a)) {
				posns[q] = true;
			}
		}
	}
	return posns
}
