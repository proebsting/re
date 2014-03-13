/*
	rxg.go -- regular expression multiple example generator

	usage:  rxg [-R] [exprfile]

	Rxg reads a set of regular expressions and synthesizes one example
	corresponding to each "accept" state in the combined DFA.

	-R	produce reproducible output by using a fixed seed

	Input is one unadorned regular expression per line.

	Output consists of the numbered input expressions followed by the
	generated examples, with the sections separated by an empty line.
	Each example is preceded by the state number and contents of the
	set of regular expressions that accept the generated string.

	Example:
	For the input
		\d+
		\d*[1-9]
		[1-9]\d*
	the output is something like:
		0. \d+
		1. \d*[1-9]
		2. [1-9]\d*

		1 { 0 } 0
		3 { 0 1 2 } 5
		2 { 0 1 } 08
		2 { 0 2 } 30

	Spring-2014 / gmt
*/
package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"rx"
	"time"
)

func main() {

	rflag := flag.Bool("R", false, "reproducible output")
	flag.Parse()

	var efile *bufio.Scanner
	if len(flag.Args()) == 0 {
		efile = rx.MkScanner("-")
	} else if len(flag.Args()) == 1 {
		efile = rx.MkScanner(flag.Args()[0])
	} else {
		log.Fatal("usage: rxg [-R] [exprfile]")
	}
	if *rflag {
		rand.Seed(0)
	} else {
		rand.Seed(int64(time.Now().Nanosecond()))
	}

	// load and process regexps
	elist := make([]string, 0)
	tlist := make([]rx.Node, 0)
	for efile.Scan() {
		line := efile.Text()
		fmt.Printf("%d. %s\n", len(elist), line)
		elist = append(elist, line)
		ptree, err := rx.Parse(line)
		rx.CkErr(err)
		tlist = append(tlist, rx.Augment(ptree, uint(len(tlist))))
	}
	rx.CkErr(efile.Err())

	// build the DFA
	dfa := rx.MultiDFA(tlist)
	fmt.Println()

	// initialize task list
	curr := Partial{dfa.Dstates[0], ""} // current partial example
	todo := make([]Partial, 0)          // list of things to do
	todo = append(todo, curr)

	// process states in breadth-first fashion
	for len(todo) > 0 { // while to-do list non-empty
		curr = todo[0] // consume one entry
		todo = todo[1:]
		ds := curr.ds
		if ds.Marked {
			continue
		}
		ds.Marked = true // mark this state as visited

		// if this is an accept state, generate output
		if ds.AccSet != nil {
			//#%#% re-follow path with different random chars?
			//#%#% would this variety be better, or worse?
			fmt.Printf("%d %s %s \n",
				ds.Index, ds.AccSet, curr.path)
		}

		// add all unmarked nodes reachable in one more step
		slist, xmap := dfa.InvertMap(curr.ds)
		alist := slist.Members()
		//%#%#% permute the list here for greater variation
		for _, arc := range alist {
			dest := dfa.Dstates[arc]
			if !dest.Marked {
				c := xmap[arc].RandChar()
				s := curr.path + string(c)
				todo = append(todo, Partial{dest, s})
			}
		}
	}
}

type Partial struct {
	ds   *rx.DFAstate // dfa state
	path string       // how we got here
}
