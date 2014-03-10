//  automata.go -- rx automata construction

package rx

import (
	"fmt"
	"io"
)

// DFA is a deterministic finite automaton.
type DFA struct {
	Leaves  []*MatchNode // leaves (positions) from parse tree
	Dstates []*DFAstate  // sets of positions
}

// DFAstate is one state in a DFA
type DFAstate struct {
	Index  uint               // index (label) of this state
	Posns  *BitSet            // set of positions in the state
	AccSet *BitSet            // set of regexs that accept here, or nil
	Dnext  map[uint]*DFAstate // transition map
}

//  DFA.Accepts returns the set of regexs that accept this string, or nil.
//  #%#% This function (only?) treats the string as Unicode runes.
func (dfa *DFA) Accepts(s string) *BitSet {
	state := dfa.Dstates[0]
	for _, r := range s {
		state = state.Dnext[uint(r)]
		if state == nil {
			return nil // unmatched char
		}
	}
	return state.AcceptBy() // end of string; in accept state?
}

//  DFA.AcceptBy returns a set of regexs that accept at this node, or nil.
func (ds *DFAstate) AcceptBy() *BitSet {
	return ds.AccSet
}

//  BuildDFA constructs an automaton from an augmented parse tree.
//  Data fields are set in the parse tree but its structure is not changed.
func BuildDFA(tree Node) *DFA {
	return MultiDFA(append(make([]Node, 0, 1), tree))
}

//  MultiDFA constructs an automoton for parallel checking of augmented trees.
func MultiDFA(tlist []Node) *DFA {
	if len(tlist) == 0 {
		panic("empty parse tree list")
	}

	// make a supertree considering all input trees as alternatives
	tree := Node(nil)
	for _, t := range tlist {
		if !IsAccept(t) && !IsAccept(t.(*ConcatNode).R) {
			panic("not an augmented parse tree")
		}
		tree = Alternate(tree, t)
	}
	dfa := &DFA{make([]*MatchNode, 0), make([]*DFAstate, 0)}

	// prepare nodes for followpos computation
	n := uint(0)
	Walk(tree, nil, func(d Node) {
		if leaf, ok := d.(*MatchNode); ok {
			// it's a leaf, so number it and remember it
			leaf.Posn = n
			n++
			dfa.Leaves = append(dfa.Leaves, leaf)
		}
		d.SetNFL() // set Nullable, FirstPos, LastPos
	})
	pmap := dfa.Leaves // map of indexes to nodes

	// compute followpos sets
	Walk(tree, nil, func(d Node) {
		d.SetFollow(pmap)
	})

	// compute DFA; see Dragon2 book p141

	// initialize first unmarked Dstate
	cs := &BitSet{}
	for _, p := range tree.Data().FirstPos.Members() {
		cs.Set(p)
	}
	dfa.Dstates = append(dfa.Dstates, &DFAstate{0, cs, nil, nil})

	// Process unmarked Dstates until none are left
	for nmarked := 0; nmarked < len(dfa.Dstates); nmarked++ {
		ds := dfa.Dstates[nmarked] // unmarked Dstate T
		ds.Dnext = make(map[uint]*DFAstate, 0)
		plist := ds.Posns.Members()     // list of p in T
		alist := validhere(pmap, plist) // potential a values
		for _, a := range alist {       // for each input symbol a
			u := followposns(pmap, plist, int(a))
			if !u.IsEmpty() {
				ustate := addstate(dfa, u) // add new state?
				ds.Dnext[a] = ustate       // register transition
			}
		}
	}

	// set the "Accept" set for each state that accepts a regex
	for _, ds := range dfa.Dstates {
		for _, px := range ds.Posns.Members() {
			p := dfa.Leaves[px]
			if IsAccept(p) {
				if ds.AccSet == nil {
					ds.AccSet = &BitSet{}
				}
				ds.AccSet.Set(p.RxIndex)
			}
		}
	}

	// return DFA
	return dfa
}

//  Addstate adds position set U to a DFA if it is distinct, returning
//  its index.  If U is not distinct, it returns the existing index.
func addstate(dfa *DFA, u *BitSet) *DFAstate {
	// start at high end because the most recent has best chance of match
	for i := len(dfa.Dstates) - 1; i >= 0; i-- {
		if dfa.Dstates[i].Posns.Equals(u) {
			return dfa.Dstates[i]
		}
	}
	// need to make a new one
	unew := &DFAstate{uint(len(dfa.Dstates)), u, nil, nil}
	dfa.Dstates = append(dfa.Dstates, unew)
	return unew
}

//  Followlist returns the union of the csets of all members of plist.
//  (This gives us fewer potential input symbols a over which to iterate.)
func validhere(pmap []*MatchNode, plist []uint) []uint {
	cs := &BitSet{}
	for _, p := range plist {
		cs = cs.Or(pmap[p].Cset)
	}
	return cs.Members()
}

//  Followposns returns the set U for a new Dstate: the set of positions
//  that are in followpos(p) for some p in plist on input symbol a.
func followposns(pmap []*MatchNode, plist []uint, a int) *BitSet {
	posns := &BitSet{}
	for _, p := range plist {
		if pmap[p].Cset.Test(uint(a)) {
			for _, q := range pmap[p].FollowPos.Members() {
				posns.Set(q)
			}
		}
	}
	return posns
}

//  InvertMap returns a list of dest states and a mapping to transition sets
func (dfa *DFA) InvertMap(ds *DFAstate) (*BitSet, map[uint]*BitSet) {
	slist := &BitSet{}
	xmap := make(map[uint]*BitSet)
	for c, ds := range ds.Dnext {
		j := ds.Index
		v := xmap[j]
		if v == nil {
			v = &BitSet{}
			xmap[j] = v
			slist.Set(j)
		}
		v.Set(c)
	}
	return slist, xmap
}

//  DumpStates prints a readable list of states.
func (dfa *DFA) DumpStates(f io.Writer) {
	for _, ds := range dfa.Dstates {

		// print index with "Accept" flag
		if ds.AcceptBy() != nil {
			fmt.Fprintf(f, "s%d# {", ds.Index)
		} else {
			fmt.Fprintf(f, "s%d. {", ds.Index)
		}

		// print position set
		for _, j := range ds.Posns.Members() {
			fmt.Fprintf(f, " p%d", j)
		}
		fmt.Fprint(f, " }")

		// invert the transition map to group input symbols by dest
		slist, xmap := dfa.InvertMap(ds)
		for _, c := range slist.Members() {
			fmt.Fprintf(f, " %s:s%d", xmap[c].Bracketed(), c)
		}
		fmt.Fprintln(f)
	}
}
