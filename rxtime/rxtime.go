/*
	rxtime.go - combine regular expressions n ways

	usage:  rxtime exprfile [n]

	Rxtime reads regular expressions from exprfile and prints statistics,
	including timings, for all possible combinations of n expressions,
	just 1 by default.

	Each output line shows, in this order:
	    Number of states in the initial combined DFA
	    Number of states in the minimized DFA
	    Time in seconds to produce the combined DFA
	    Time in seconds to minimize the DFA
	    If n == 1, the computed "complexity score"
	    Ordinals of the expressions combined in this DFA

	Erroneous expressions are silently ignored.

	spring 2014 / gmt
*/
package main

import (
	"fmt"
	"log"
	"os"
	"rx"
	"rx/rxsys"
	"strconv"
)

func main() {
	// get command line options
	nways := 1
	if len(os.Args) == 3 {
		nways, _ = strconv.Atoi(os.Args[2])
	}
	if len(os.Args) < 2 || len(os.Args) > 3 || nways < 1 {
		log.Fatal("usage: rxtime exprfile [n]")
	}
	filename := os.Args[1]

	// load expressions from file
	exprs := rx.LoadExpressions(filename, nil)
	nexprs := len(exprs)
	if nways < 1 || nways > nexprs {
		log.Fatal(fmt.Sprintf(
			"cannot combine %d expressions(s) in %d way(s)",
			nexprs, nways))
	}

	// record individual complexity scores and make augmented parse trees
	cx := make([]int, nexprs)
	for i, t := range exprs {
		cx[i] = rx.ComplexityScore(t.Tree)
		t.Tree = rx.Augment(t.Tree, i)
	}

	// initialize index list for first combination {0,1,2...}
	xlist := make([]int, nways)
	for i := range xlist {
		xlist[i] = i
	}

	// try all possible n-way combinations by varying the index list
	tlist := make([]rx.Node, nways)
	for xlist != nil {
		for i, x := range xlist {
			tlist[i] = exprs[x].Tree
		}
		_ = rxsys.Interval()             // reset timer
		dfa1 := rx.MultiDFA(tlist)       // make DFA
		t1 := rxsys.Interval().Seconds() // measure time
		dfa2 := dfa1.Minimize()          // minimize DFA
		t2 := rxsys.Interval().Seconds() // measure time
		fmt.Printf("%6d %6d %8.3f %8.3f",
			len(dfa1.Dstates), len(dfa2.Dstates), t1, t2)
		if nways == 1 {
			fmt.Printf(" %6d", cx[xlist[0]])
		}
		fmt.Print("   {")
		for _, x := range xlist {
			fmt.Printf(" %d", x)
		}
		fmt.Print(" }\n")
		xlist = advance(xlist, nexprs) // get next combination
	}
}

// advance an index list (initially {0,1,2...}) to next combination in sequence
// (n.b. although this returns a slice, it is changing the underlying array)
func advance(xlist []int, nitems int) []int {
	nchoose := len(xlist)
	i := nchoose - 1
	// find an index that can be incremented
	for i >= 0 && xlist[i] > (nitems-(nchoose-i)-1) {
		i--
	}
	if i < 0 {
		return nil // no more combinations to try
	}
	// increment index i and reset all that follow
	xlist[i]++
	for i++; i < nchoose; i++ {
		xlist[i] = xlist[i-1] + 1
	}
	return xlist
}
