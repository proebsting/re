//  rx.go -- some top-level entry points for the rx library package

//  Rx provides facilities for dealing with regular expressions.
package rx

import (
	"fmt"
)

var _ = fmt.Printf //#%#%#% for debugging

//  Match tests whether a string is matched by a regular expression.
func Match(rexpr string, s string) (bool, error) {
	dfa, err := Compile(rexpr)
	if err != nil {
		return false, err
	}
	return (dfa.Accepts(s) != nil), nil
}

//  Compile makes a minimized DFA from a regular expression.
//  The DFA can be reused for multiple matches, saving compilation costs.
func Compile(rexpr string) (*DFA, error) {
	ptree, err := Parse(rexpr) // make faithful parse tree
	if err != nil {
		return nil, err
	}
	atree := Augment(ptree, 0) // make augmented parse tree
	dfa := BuildDFA(atree)     // build DFA from augmented tree
	dfa = dfa.Minimize()       // minimize the number of states
	return dfa, nil            // return it

}

//  DFA.Accepts returns the set of regexps that accept a string, or nil.
//  This function (#%#% only?) treats the input string as Unicode runes.
func (dfa *DFA) Accepts(s string) *BitSet {
	state := dfa.Dstates[0]
	for _, r := range s {
		state = state.Dnext[uint(r)]
		if state == nil {
			return nil // unmatched char
		}
	}
	return state.AccSet // end of string; return accepting set if any
}

//  Augment produces a modified parse tree in preparation for building a DFA.
//  The original tree is modified in place, and a new root is returned.
//  This new root concatenates an Accept node to the input tree.
//  Additionally, any fixed {m,n} replications with m>1 or n>1 are replaced
//  by concatenations of duplicated subtrees.
func Augment(tree Node, rxindex uint) Node {
	Walk(tree, nil, func(d Node) {
		switch d.(type) {
		case *ConcatNode:
			cnode := d.(*ConcatNode)
			cnode.L = replfix(cnode.L)
			cnode.R = replfix(cnode.R)
		case *AltNode:
			anode := d.(*AltNode)
			alts := make([]Node, 0, len(anode.Alts))
			for _, e := range anode.Alts {
				alts = append(alts, replfix(e))
			}
			anode.Alts = alts
		case *ReplNode:
			rnode := d.(*ReplNode)
			rnode.Child = replfix(rnode.Child)
		case *MatchNode:

		}
	})
	tree = replfix(tree) // need to handle top node, too
	return Concatenate(tree, Accept(rxindex))
}
