/*
	rxr.go -- regular expression reader

	#%#% This program has been superseded by rxplor.

	usage:  rxr [-Q] [-R] [exprfile]

		-Q	disable printing of automata details
		-R	produce reproducible output for testing purposes

	Rxr reads regular expressions and illustrates the workings of the
	rx package.  For each expression, rxr prints:
		- the original regular expression as read
		- the expression expressed by the parse tree
		- the minimum and maximum lengths of matching strings
		- synthetic examples with varied limits on replication
		- the parse tree nodes labeled with {nullable, first, last}
		- the positions (leaves) of the parse tree with followsets
		- the states of the resulting DFA with transitions

	Regular expressions are read one per line from exprfile.
	A line beginning with '#', or an empty line, is treated as a comment.

	The output is informative but not rigorous.  Ambiguities may arise
	if pattern metacharacters are used as ordinary matching characters.

	Each synthesized example is validated by the DFA; a rejected string
	prints the word "[FAIL]" and signifies a program bug.

	Spring-2014 / gmt
*/
package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"rx"
	"time"
	"unicode/utf8"
)

func main() {
	qflag := flag.Bool("Q", false, "disable automata printing")
	rflag := flag.Bool("R", false, "reset ranseed before each rx")
	rand.Seed(int64(time.Now().Nanosecond()))

	// load and process regexps
	rx.LoadExpressions(rx.OneInputFile(), func(in *rx.RegExParsed) {
		spec := in.Expr
		if !in.IsExpr() {
			if in.Meta == nil { // echo if not metadata
				fmt.Printf("\n%s\n", spec)
			}
			return
		}
		// print accumulated metadata
		in.ShowMeta(os.Stdout, "")
		// print regexp
		fmt.Printf("\nregexp: %s\n", spec)
		if in.Err != nil {
			fmt.Println("ERROR: ", in.Err)
			return
		}
		t := in.Tree
		fmt.Printf("tree:   %v\n", t)
		augt := rx.Augment(t, 0) // make augmented tree
		if !*qflag {             // if -Q not given, print it
			fmt.Printf("augmnt: %v\n", augt)
		}

		fmt.Printf("length: %d to ", t.MinLen())
		x := t.MaxLen()
		if x >= 0 {
			fmt.Println(x)
		} else {
			fmt.Println("*")
		}

		dfa := rx.BuildDFA(augt) // make DFA from augmented tree

		if *rflag { // if -R given, reset random seed
			rand.Seed(0) // for independent, reproducible output
		}
		rx.ShowLabel(os.Stdout, "Examples")
		examples(dfa, t, 0) // gen and test w/ max repl of 0
		examples(dfa, t, 1) // ... and 1
		examples(dfa, t, 2) // ... and 2
		examples(dfa, t, 3) // ... and 3
		examples(dfa, t, 5) // ... and 5
		examples(dfa, t, 8) // ... and 8

		if !*qflag { // if -Q not given, print automata
			// show tree details incl nullable, firstpos, lastpos
			dfa.ShowTree(os.Stdout, augt, "Annotated Tree")
			// dump positions and follow sets
			dfa.ShowNFA(os.Stdout, "NFA")
			// dump DFA
			dfa.ShowStates(os.Stdout, "Initial DFA")
			// generate minimal DFA and dump that
			dfa = dfa.Minimize()
			dfa.ShowStates(os.Stdout, "Minimized DFA")
		}
	})
}

const linemax = 79

//   examples generates a line's worth of examples with max replication n.
func examples(dfa *rx.DFA, tree rx.Node, n int) {
	s := fmt.Sprintf("ex(%d):", n)
	nprinted := 0
	fmt.Print(s)
	ncolm := utf8.RuneCountInString(s)
	for {
		s := rx.Specimen(tree, n)
		t := rx.Protect(s)
		ncolm += 2 + utf8.RuneCountInString(t)
		if nprinted > 0 && ncolm > linemax {
			break
		}
		fmt.Printf("  %s", t)
		if dfa.Accepts(s) == nil {
			fmt.Print(" [FAIL]")
			ncolm += 7
		}
		nprinted++
	}
	fmt.Println()
}
