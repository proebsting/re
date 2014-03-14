/*
	rxx.go -- regular expression cross product with candidate list

	usage:  rxx efile sfile

	Rxx reads a set of up to 61 regular expressions, one per line,
	from efile.  It then tests every line from sfile against each
	regular expression, printing a grid of results on standard output.

	A line beginning with '#', or an empty line, is treated as a comment.

	Spring-2014 / gmt
*/
package main

import (
	"fmt"
	"log"
	"os"
	"rx"
)

type tester struct { // one regular expression for testing
	label string  // one-character label
	spec  string  // regular expression specification
	tree  rx.Node // unaugmented parse tree
	index int     // result index
}

func main() {
	args := os.Args
	if len(args) != 3 {
		log.Fatal("usage: rxx efile sfile")
	}
	efile := rx.MkScanner(args[1])
	sfile := rx.MkScanner(args[2])

	labels :=
		"123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	elist := make([]tester, 0, len(labels))

	// load and compile regexps
	fmt.Println()
	tlist := make([]rx.Node, 0) // list of valid parse trees
	for efile.Scan() {
		spec := efile.Text()
		var ptree rx.Node
		var err error
		if rx.IsComment(spec) {
			ptree, err = nil, nil
		} else {
			ptree, err = rx.Parse(spec)
		}
		if err != nil {
			fmt.Printf("ERR %s\n", spec)
			elist = append(elist, tester{" ", spec, nil, 0})
		} else if ptree == nil {
			fmt.Printf("    %s\n", spec)
			elist = append(elist, tester{" ", spec, nil, 0})
		} else {
			i := len(tlist)
			if i >= len(labels) {
				log.Fatal("too many regular expressions")
			}
			label := string(labels[i : i+1])
			fmt.Printf("%s:  %s\n", label, spec)
			atree := rx.Augment(ptree, uint(len(tlist)))
			tlist = append(tlist, atree)
			elist = append(elist, tester{label, spec, ptree, i})
		}
	}
	rx.CkErr(efile.Err())
	dfa := rx.MultiDFA(tlist)

	// read and test candidate strings
	fmt.Println()
	for sfile.Scan() {
		s := string(sfile.Bytes())
		results := dfa.Accepts(s)
		if results == nil {
			results = &rx.BitSet{}
		}
		for _, e := range elist {
			if e.tree == nil {
				fmt.Print(" ")
			} else {
				if results.Test(uint(e.index)) {
					fmt.Print(e.label)
				} else {
					fmt.Print("-")
				}
			}
		}
		fmt.Printf("  %s\n", s)
	}
	rx.CkErr(sfile.Err())
}
