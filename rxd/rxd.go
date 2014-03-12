/*
	rxd.go -- regular expression to Dot

	usage:  rxd ["rexpr"...]

	Rxd builds a DFA that accepts one or more regular expressions
	and generates Dot (Graphviz) directives for displaying the graph.

	If no expressions are given as command arguments, rxd reads
	expressions from standard input, one per line, with '#'
	indicating a comment line.

	Spring-2014 / gmt
*/
package main

import (
	"fmt"
	"os"
	"rx"
)

func main() {
	// build a list of regular espressions
	elist := make([]string, 0)
	for i := 1; i < len(os.Args); i++ {
		elist = append(elist, os.Args[i])
	}
	if len(elist) == 0 {
		ifile := rx.MkScanner("-")
		for ifile.Scan() {
			e := ifile.Text()
			if !rx.IsComment(e) {
				elist = append(elist, ifile.Text())
			}
		}
		rx.CkErr(ifile.Err())
	}
	// build a list of augmented parse trees
	tlist := make([]rx.Node, 0, len(elist))
	for _, e := range elist {
		ptree, err := rx.Parse(e)
		rx.CkErr(err)
		tlist = append(tlist, rx.Augment(ptree, uint(len(tlist))))
	}
	// build the DFA
	dfa := rx.MultiDFA(tlist)
	// generate Dot output
	var label string
	if len(tlist) > 1 {
		label = fmt.Sprintf("(%d expressions)", len(tlist))
	} else {
		label = elist[0]
	}
	dfa.ToDot(os.Stdout, label)
}
