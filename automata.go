//  automata.go -- deterministic finite state automata construction

package rx

import (
	"fmt"
	"io"
	"strconv"
)

//  A DFA is a deterministic finite automaton for matching regular expressions.
//  The first entry in the Dstates array is the start state.
type DFA struct {
	Tree     Node         // final compiled augmented parse tree
	Leaves   []*MatchNode // leaves (positions) from parse tree
	Dstates  []*DFAstate  // sets of positions
	PartList []*Partition // partition list during minimization
}

//  newDFA makes a new, empty DFA.
func newDFA(tree Node) *DFA {
	return &DFA{tree, make([]*MatchNode, 0), make([]*DFAstate, 0), nil}
}

//  DFAstate is one state in a DFA.
//  The Dnext table maps input symbols to successor states.
type DFAstate struct {
	Index   uint               // index (label) of this state
	Marked  bool               // true when "marked" during traversal
	Posns   *BitSet            // set of positions in the state
	AccSet  *BitSet            // set of regexps that accept here, or nil
	Dnext   map[uint]*DFAstate // transition map
	PartNum uint               // partition number during minimization
}

//  DFA.newState makes a new DFAstate and adds it to a DFA.
func (dfa *DFA) newState(posns *BitSet) *DFAstate {
	ds := &DFAstate{uint(len(dfa.Dstates)), false, posns, nil,
		make(map[uint]*DFAstate, 0), 0}
	dfa.Dstates = append(dfa.Dstates, ds)
	return ds
}

//  BuildDFA constructs an automaton from an augmented parse tree.
//  Data fields are set in the parse tree but its structure is not changed.
func BuildDFA(tree Node) *DFA {
	return MultiDFA(append(make([]Node, 0, 1), tree))
}

//  MultiDFA constructs an automaton for parallel checking of augmented trees.
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
	dfa := newDFA(tree)

	// prepare nodes for followpos computation
	n := uint(0)
	Walk(tree, nil, func(d Node) {
		if leaf, ok := d.(*MatchNode); ok {
			// it's a leaf, so number it and remember it
			leaf.Posn = n
			n++
			dfa.Leaves = append(dfa.Leaves, leaf)
		}
		d.SetNFL()                     // Nullable, FirstPos, LastPos
		d.Data().FollowPos = &BitSet{} // init empty FollowPos
	})
	pmap := dfa.Leaves // map of indexes to nodes

	// compute followpos sets
	Walk(tree, nil, func(d Node) {
		d.SetFollow(pmap)
	})

	// compute DFA: see Dragon2 book p141

	// initialize first unmarked Dstate
	posns := &BitSet{}
	for _, p := range tree.Data().FirstPos.Members() {
		posns.Set(p)
	}
	knownStates := make(map[string]*DFAstate) // make hashtable of states
	dfa.lookup(knownStates, posns)            // add initial state

	// process unmarked Dstates until none are left
	for nmarked := 0; nmarked < len(dfa.Dstates); nmarked++ {
		ds := dfa.Dstates[nmarked]               // next unmkd Dstate T
		plist := ds.Posns.Members()              // list of p in T
		for _, a := range dfa.validHere(plist) { // potential input syms
			u := dfa.followSet(plist, a) // set of follow posns
			if !u.IsEmpty() {
				// find or make a matching state
				ustate := dfa.lookup(knownStates, u)
				ds.Dnext[a] = ustate // transition on a
			}
		}
	}

	// set the accepting set for each state that accepts a regexp
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

//  lookup finds a state matching position set U in map m.
//  If U is distinct from existing entries, a new state is created.
func (dfa *DFA) lookup(m map[string]*DFAstate, u *BitSet) *DFAstate {
	k := u.Key()
	ds := m[k]
	if ds == nil {
		ds = dfa.newState(u) // need to make a new one
		m[k] = ds
	}
	return ds
}

//  validHere returns the union of the csets of all members of plist.
//  (This gives us fewer potential input symbols a over which to iterate.)
func (dfa *DFA) validHere(plist []uint) []uint {
	pmap := dfa.Leaves
	cs := &BitSet{}
	for _, p := range plist {
		cs.OrWith(pmap[p].Cset)
	}
	return cs.Members()
}

//  followSet returns those positions in followpos(p) where p matches a
func (dfa *DFA) followSet(plist []uint, a uint) *BitSet {
	fset := &BitSet{}
	for _, p := range plist {
		leaf := dfa.Leaves[p]
		if leaf.Cset.Test(a) {
			fset.OrWith(leaf.FollowPos)
		}
	}
	return fset
}

//  DFAstate.InvertMap lists dest states and maps to transition sets.
//  The list duplicates the map indexes but is more easily traversed in order.
func (ds *DFAstate) InvertMap() (*BitSet, map[uint]*BitSet) {
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

//  DFA.ShowNFA prints the positions and followsets.
func (dfa *DFA) ShowNFA(f io.Writer) {
	fmt.Fprintf(f, "begin => %s\n", dfa.Tree.Data().FirstPos)
	for _, m := range dfa.Leaves {
		fmt.Fprintf(f, "p%d. %s => %s\n", m.Posn, m, m.FollowPos)
	}
}

//  DFA.DumpStates prints a readable list of states.
func (dfa *DFA) DumpStates(f io.Writer) {
	for _, ds := range dfa.Dstates {
		// print partition index
		//#%#% fmt.Fprintf(f, "[%d] ", ds.PartNum)

		// print index with "accept" flag '#'
		if ds.AccSet != nil {
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
		slist, xmap := ds.InvertMap()
		for _, c := range slist.Members() {
			fmt.Fprintf(f, " %s:s%d", xmap[c].Bracketed(), c)
		}
		fmt.Fprintln(f)
	}
}

//  DFA.ToDot generates a Dot (GraphViz) representation of the DFA.
func (dfa *DFA) ToDot(f io.Writer, label string) {
	fmt.Fprintln(f, "//", label)
	fmt.Fprintln(f, "digraph DFA {")
	fmt.Fprintf(f, "label=%s\n", strconv.Quote(label))
	fmt.Fprintln(f, "node [shape=circle, height=.3, margin=0, fontsize=10]")
	fmt.Fprintln(f, "s0 [shape=triangle, regular=true]")
	for _, src := range dfa.Dstates {
		if src.AccSet != nil {
			fmt.Fprintf(f, "s%d[shape=doublecircle]\n", src.Index)
		}
		slist, xmap := src.InvertMap()
		for _, dst := range slist.Members() {
			fmt.Fprintf(f, "s%d->s%d[label=\"%s\"]\n",
				src.Index, dst, xmap[uint(dst)].Bracketed())
		}
	}
	fmt.Fprintln(f, "}")
}
