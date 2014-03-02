/*
	rxx.go -- regular expression cross product with candidate list

	usage:  recross efile sfile

	Rexp reads a set of regular expressions, one per line, from efile.
	It then tests every line from sfile against each regular expression,
	printing the results on standard output.
*/

package main

import (
	"fmt"
	"log"
	"os"
	"rx"
)

type tester struct {
	label string
	spec  string
	dfa   *rx.DFA
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
	for i := 0; efile.Scan(); i++ {
		label := string(labels[i : i+1])
		spec := efile.Text()
		dfa, _ := rx.Compile(spec)
		re := tester{label, spec, dfa}
		if dfa == nil {
			fmt.Printf("ERR %s\n", spec)
		} else {
			fmt.Printf("%s:  %s\n", label, spec)
		}
		elist = append(elist, re)
	}
	rx.CkErr(efile.Err())

	// read and test candidate strings
	fmt.Println()
	for sfile.Scan() {
		s := string(sfile.Bytes())
		for _, e := range elist {
			if e.dfa == nil {
				fmt.Print(" ")
			} else {
				if e.dfa.Accepts(s) {
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
