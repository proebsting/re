//  automata.go -- rx automata construction

//#%#% uint16 should be a named type

//  Rx provides facilities for dealing with regular expressions.
package rx

import (
	"fmt"
)

var _ = fmt.Printf //#%#%#% for debugging

// DFA is a deterministic finite automaton.
type DFA struct {
	Leaves  []*MatchNode // leaves (positions) from parse tree
	Dstates []*DFAstate  // sets of positions
}

// DFAstate is one state in a DFA
type DFAstate struct {
	Index int                  // index (label) of this state
	Posns *Cset                // set of positions in the state
	Dnext map[uint16]*DFAstate // transition map
}

// n.b. used a Cset for Posns to allow easy comparison of position sets.

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
			leaf.Posn = n
			n++ // number the leaf nodes from zero
			dfa.Leaves = append(dfa.Leaves, leaf)
		}
	})
	pmap := dfa.Leaves // map of indexes to nodes

	// compute followpos sets
	tree.Walk(nil, func(d Node) {
		d.SetFollow()
	})

	// compute DFA; see Dragon2 book p141

	// initialize first unmarked Dstate
	cs := CharSet("") //#%#% ugh(name)
	for p := range tree.Data().FirstPos {
		cs.Set(uint(p.Posn))
	}
	dfa.Dstates = append(dfa.Dstates, &DFAstate{0, cs, nil})

	// Process unmarked Dstates until none are left
	for nmarked := 0; nmarked < len(dfa.Dstates); nmarked++ {
		d := dfa.Dstates[nmarked] // unmarked Dstate T
		//#%#% fmt.Printf("s%d: i%d %s\n",nmarked,d.Index,d.Posns)
		d.Dnext = make(map[uint16]*DFAstate, 0)
		plist := d.Posns.Members() // list of p in T
		//#%#% fmt.Printf("  pl: %v\n",plist)
		alist := validhere(pmap, plist) // potential a values
		//#%#% fmt.Printf("  al: %v\n",alist)
		for _, a := range alist { // for each input symbol a
			u := followposns(pmap, plist, int(a))
			//#%#% fmt.Printf("    ch %s: %v\n", string(a), u.Members())
			if !u.IsEmpty() {
				ustate := addstate(dfa, u) // add new state?
				d.Dnext[a] = ustate        // register transition
				//#%#% fmt.Printf("	    state %d\n",ustate.Index)
			}
		}
	}

	// return DFA and augmented tree
	return dfa, tree
}

//  Addstate adds position set U to a DFA if it is distinct, returning
//  its index.  If U is not distinct, it returns the existing index.
func addstate(dfa *DFA, u *Cset) *DFAstate {
	// start at high end because the most recent has best chance of match
	for i := len(dfa.Dstates) - 1; i >= 0; i-- {
		if dfa.Dstates[i].Posns.Equals(u) {
			return dfa.Dstates[i]
		}
	}
	// need to make a new one
	unew := &DFAstate{len(dfa.Dstates), u, nil}
	dfa.Dstates = append(dfa.Dstates, unew)
	return unew
}

//  Followlist returns the union of the csets of all members of plist.
//  (This gives us fewer potential input symbols a over which to iterate.)
func validhere(pmap []*MatchNode, plist []uint16) []uint16 {
	cs := CharSet("")
	for _, p := range plist {
		cs = cs.Or(pmap[p].cset)
	}
	return cs.Members()
}

//  Followposns returns the set U for a new Dstate: the set of positions
//  that are in followpos(p) for some p in plist on input symbol a.
func followposns(pmap []*MatchNode, plist []uint16, a int) *Cset {
	posns := CharSet("")
	for _, p := range plist {
		if pmap[p].cset.Test(uint(a)) {
			for q := range pmap[p].FollowPos {
				posns.Set(uint(q.Posn))
			}
		}
	}
	return posns
}