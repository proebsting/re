//  minimize.go -- minimization of a DFA

package rx

import (
	"fmt"
)

var _ = fmt.Printf //#%#% for debugging

var deadstate *DFAstate // "dead state" used during minimization

//  A Partition is a subset of the states of a DFA.
type Partition struct {
	Automata *DFA      // the containing DFA
	Index    uint      // index (label) of this partition
	StateSet *BitSet   // members of this partition
	NewState *DFAstate // destination state in minimized DFA
}

//  DFA.newPartition creates a new partition and adds it to a DFA.
func (dfa *DFA) newPartition() *Partition {
	p := &Partition{dfa, uint(len(dfa.PartList)), &BitSet{}, nil}
	dfa.PartList = append(dfa.PartList, p)
	return p
}

//  DFA.Minimize returns an optimized variant with a minimum number of states.
//  The original DFA remains valid although some of its data structures have
//  been disturbed in the process.
func (dfa *DFA) Minimize() *DFA {

	// add and remember a "dead state" with no exits
	deadstate = dfa.newState(&BitSet{})

	// make a partition list with one initial partition
	// give this partition the accepting set of the start state
	// so the resulting DFA will also have the start state first
	dfa.PartList = make([]*Partition, 0, len(dfa.Dstates)+1)
	p := dfa.newPartition()
	amap := make(map[*BitSet]*Partition)
	amap[dfa.Dstates[0].AccSet] = p

	// insert states into partitions;
	// make a new partition for each distinct set of accepting states seen
	for _, ds := range dfa.Dstates {
		ds.PartNum = 0    // clear old partition number
		pset := ds.AccSet // copy accepting set
		p := amap[pset]   // find partition for this set
		if p == nil {     // if there isn't one yet, make one
			p = dfa.newPartition()
			amap[pset] = p
		}
		p.insert(ds) // move state to partition
	}

	// repeatedly subdivide partitions until can do so no more
	nchanged := 1
	for nchanged != 0 {
		// fmt.Println("[partitioning:]") //#%#%#% TEMP // once per pass
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
	// fmt.Println("[merging:]") //#%#%#% TEMP
	minim := newDFA(dfa.Tree)
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

	// return the new minimized DFA
	return minim
}

//  Partition.insert moves a state into a partition.
func (p *Partition) insert(ds *DFAstate) {
	dfa := p.Automata
	i := ds.Index
	old := ds.PartNum
	dfa.PartList[old].StateSet.Clear(i)
	p.StateSet.Set(i)
	ds.PartNum = p.Index
	// fmt.Printf("move s%d : ptn %d to %d\n", i, old, p.Index) //#%#% TEMP
}

//  Partition.distinguish returns an input for partitioning, or -1 for none.
//  The return value is any input that maps states into two or more groups.
func (p *Partition) distinguish() int {
	mlist := p.StateSet.Members()
	for i := range mlist {
		for j := i + 1; j < len(mlist); j++ {
			s1 := p.Automata.Dstates[mlist[i]]
			s2 := p.Automata.Dstates[mlist[j]]
			// check every input of s1 against s2 and vice versa;
			// this makes duplicate checks, but is simpler and
			// possibly cheaper than avoiding the duplication
			x := s1.distinguish(s2)
			if x >= 0 {
				return x
			}
			x = s2.distinguish(s1)
			if x >= 0 {
				return x
			}
		}
	}
	return -1
}

//  DFAstate.distinguish checks every recognized input against another state,
//  returning an input that leads to a different partition or -1 if there is
//  none.
func (s1 *DFAstate) distinguish(s2 *DFAstate) int {
	for x := range s1.Dnext {
		if s1.partOn(int(x)) != s2.partOn(int(x)) {
			return int(x)
		}
	}
	return -1
}

//  Partition.divide repartitions this partition based on input x.
func (p *Partition) divideBy(x int) {
	// fmt.Printf("[partition %d]\n", p.Index) //#%#% TEMP
	// get a list of partition members
	// all distinguishable from the first by x go into a new partition
	dfa := p.Automata              // DFA being partitioned
	mlist := p.StateSet.Members()  // list of partition members
	ds := dfa.Dstates[mlist[0]]    // first member
	pi := ds.partOn(x)             // partition reached on input x
	q := p.Automata.newPartition() // new partition for all who differ
	for i := 1; i < len(mlist); i++ {
		ds = dfa.Dstates[mlist[i]] // candidate state
		if ds.partOn(x) != pi {
			// fmt.Printf("%#v: ", string(x)) //#%#% TEMP
			q.insert(ds)
		}
	}
}

//  DFAstate.partOn returns the index of the partition reached by input x.
func (ds *DFAstate) partOn(x int) uint {
	dd := ds.Dnext[uint(x)]
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
			odest := os.Dnext[uint(x)] // get old dest state
			ndest := dfa.PartList[odest.PartNum].NewState
			ds.Dnext[uint(x)] = ndest // set new destination
		}
	}
}
