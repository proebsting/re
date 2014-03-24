//  rx.go -- top-level entry points

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

//  Augment produces a modified parse tree in preparation for building a DFA.
//  The original tree is modified in place, and a new root is returned.
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
