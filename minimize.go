//  minimize.go -- minimization of a DFA

package rx

import (
	"fmt"
)

var deadstate *DFAstate // "dead state" used during minimization

var DBG_MIN = false // enable/disable minimization debugging output

//  A Partition is a subset of the states of a DFA.
type Partition struct {
	Automata *DFA      // the containing DFA
	Index    int       // index (label) of this partition
	StateSet *BitSet   // members of this partition
	NewState *DFAstate // destination state in minimized DFA
}

//  DFA.newPartition creates a new partition and adds it to a DFA.
func (dfa *DFA) newPartition() *Partition {
	p := &Partition{dfa, len(dfa.PartList), &BitSet{}, nil}
	dfa.PartList = append(dfa.PartList, p)
	return p
}

//  DFA.Minimize returns an optimized variant with a minimum number of states.
//  The original DFA remains valid although some of its data structures have
//  been disturbed in the process.
func (dfa *DFA) Minimize() *DFA {
	// preprocess states to create minimal "input lists"
	wset := dfa.Witnesses()
	for _, ds := range dfa.Dstates {
		ds.setinputs(wset)
	}

	// add and remember a "dead state" with no exits
	deadstate = dfa.newState(&BitSet{})

	// make a partition list with one initial partition
	// give this partition the accepting set of the start state
	// so the resulting DFA will also have the start state first
	dfa.PartList = make([]*Partition, 0, len(dfa.Dstates)+1)
	p := dfa.newPartition()
	amap := make(map[string]*Partition)
	amap[dfa.Dstates[0].AccSet.Key()] = p

	// insert states into partitions;
	// make a new partition for each distinct set of accepting states seen
	for _, ds := range dfa.Dstates {
		ds.PartNum = 0          // clear old partition number
		pset := ds.AccSet.Key() // get key from accepting set
		p := amap[pset]         // find partition for this set
		if p == nil {           // if there isn't one yet, make one
			p = dfa.newPartition()
			amap[pset] = p
		}
		p.insert(ds) // move state to partition
	}

	// repeatedly subdivide partitions until can do so no more
	nchanged := 1
	for nchanged != 0 {
		if DBG_MIN {
			fmt.Println("[partitioning:]") // once per pass
		}
		nchanged = 0
		// loop manually (vs "range" expr) to catch new partns as added
		for i := 0; i < len(dfa.PartList); i++ {
			p := dfa.PartList[i]
			for { // until distinguish() fails
				x := p.distinguish()
				if x < 0 {
					break
				}
				p.divideBy(x)
				nchanged++
			}
		}
	}

	// make a new DFA with one state for each partition,
	// but exclude the partition containing the dead state
	if DBG_MIN {
		fmt.Println("[merging:]")
	}
	minim := newDFA(dfa.Tree)
	minim.Leaves = dfa.Leaves
	for _, p = range dfa.PartList {
		if p.Index != deadstate.PartNum {
			p.NewState = minim.newState(&BitSet{})
		}
	}

	// now that all the new states exist, fill in their contents
	for _, p = range dfa.PartList {
		if p.NewState != nil {
			p.NewState.mergeFrom(p)
		}
	}

	// remove the dead state we added temporarily to the old DFA
	dfa.Dstates = dfa.Dstates[:len(dfa.Dstates)-1]

	// reclaim partition list and input-list memory
	dfa.PartList = nil
	for _, ds := range dfa.Dstates {
		ds.InpList = nil
	}

	// return the new minimized DFA
	return minim
}

//  DFAstate.setinputs computes representative inputs for each state
func (ds *DFAstate) setinputs(witnesses *BitSet) {
	cset := CharSet("")
	for x := range ds.Dnext {
		cset.Set(x)
	}
	ds.InpList = (cset.AndWith(witnesses)).Members()
}

//  Partition.insert moves a state into a partition.
func (p *Partition) insert(ds *DFAstate) {
	dfa := p.Automata
	i := ds.Index
	old := ds.PartNum
	dfa.PartList[old].StateSet.Clear(i)
	p.StateSet.Set(i)
	ds.PartNum = p.Index
	if DBG_MIN {
		fmt.Printf("move s%d : ptn %d to %d\n", i, old, p.Index)
	}
}

//  Partition.distinguish returns an input for partitioning, or -1 for none.
//  The return value is any input that maps states into two or more groups.
func (p *Partition) distinguish() int {
	psnums := p.StateSet.Members() // index numbers of states in partition
	for _, pn1 := range psnums {
		s1 := p.Automata.Dstates[pn1]
		for _, pn2 := range psnums {
			s2 := p.Automata.Dstates[pn2]
			// check representative inputs of s1 against s2
			// (and vice versa when they switch roles)
			for _, x := range s1.InpList {
				if s1.partOn(x) != s2.partOn(x) {
					return x
				}
			}
		}
	}
	return -1
}

//  Partition.divide repartitions this partition based on input x.
func (p *Partition) divideBy(x int) {
	if DBG_MIN {
		fmt.Printf("[partition %d by %#v]\n", p.Index, string(x))
	}
	// get a list of partition members
	// and note all that are distinguishable from the first by input x
	dfa := p.Automata             // DFA being partitioned
	mlist := p.StateSet.Members() // list of partition members
	tomove := &BitSet{}           // set of those to move
	d0 := dfa.Dstates[mlist[0]]   // first member
	for i := 1; i < len(mlist); i++ {
		ds := dfa.Dstates[mlist[i]] // candidate state
		if ds.partOn(x) != d0.partOn(x) {
			tomove.Set(ds.Index)
		}
	}
	// now (all at once) move the distinguished states to a new partition
	q := p.Automata.newPartition() // new partition for all who differ
	for _, i := range tomove.Members() {
		q.insert(dfa.Dstates[i])
	}
}

//  DFAstate.partOn returns the index of the partition reached by input x.
func (ds *DFAstate) partOn(x int) int {
	dd := ds.Dnext[x]
	if dd == nil {
		dd = deadstate
	}
	return dd.PartNum
}

//  DFAstate.mergeFrom makes a merged state in a new DFA from an old partition.
func (ds *DFAstate) mergeFrom(p *Partition) {
	ds.PartNum = p.Index                     // just for some history
	dfa := p.Automata                        // the old DFA
	for _, i := range p.StateSet.Members() { // for each partition member
		os := dfa.Dstates[i]      // get old state
		ds.Posns.OrWith(os.Posns) // merge positions
		if ds.AccSet == nil {
			ds.AccSet = os.AccSet // first seen, or just nil
		} else {
			ds.AccSet.OrWith(os.AccSet) // merge acceptors
		}
		for x := range os.Dnext { // for each input
			odest := os.Dnext[x] // get old dest state
			ndest := dfa.PartList[odest.PartNum].NewState
			ds.Dnext[x] = ndest // set new destination
		}
	}
}
