//  synth.go -- code for generating examples of matching expressions

package rx

//  A DFAexample is a matching string synthesized from a DFA.
type DFAexample struct {
	State   uint   // index of accepting state in DFA
	RXset   []uint // set of matching regular expression indexes
	Example string // example string
}

//  A partx is a partially built example used on the task list.
type partx struct {
	ds   *DFAstate // DFA state pointer
	path string    // how we got here
}

//  DFA.Synthesize produces a set of examples base on a DFA.
//  One example is produced for each distinct accepting state in the DFA.
//  Each example indicates which accepting state was reached and gives the
//  set of regular expressions matched at that point (by their indexes).
func (dfa *DFA) Synthesize() []DFAexample {

	// initialize the task list
	todo := make([]partx, 0) // list of things to do
	todo = append(todo, partx{dfa.Dstates[0], ""})

	// process states in breadth-first fashion
	results := make([]DFAexample, 0)
	for len(todo) > 0 { // while to-do list non-empty
		curr := todo[0]
		todo = todo[1:] // consume one entry
		ds := curr.ds
		if ds.Marked { // if we've already been here
			continue
		}
		ds.Marked = true // mark this state as visited

		// if this is an accepting state, note the result
		if ds.AccSet != nil {
			// we could re-follow the path to change random chars,
			// but would that variety be better, or worse?
			results = append(results, DFAexample{
				ds.Index, ds.AccSet.Members(), curr.path})
		}

		// add all unmarked nodes reachable in one more step
		slist, xmap := curr.ds.InvertMap()
		alist := slist.Members()
		// for greater randomness we could produce the list here,
		// but again, would that be better -- or worse?
		for _, arc := range alist {
			dest := dfa.Dstates[arc]
			if !dest.Marked {
				c := xmap[arc].RandChar()
				s := curr.path + string(c)
				todo = append(todo, partx{dest, s})
			}
		}
	}

	return results
}
