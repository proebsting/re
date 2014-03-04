/*
	rxr.go -- regular expression reader

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
	A line beginning with '#' is treated as a comment.

	The output is informative but not rigorous.  Ambiguities may arise
	if pattern metacharacters are used as ordinary matching characters.

	Each synthesized example is validated by the DFA; a rejected string
	prints the word "[FAIL]" and signifies a program bug.

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
	qflag := flag.Bool("Q", false, "disable automata printing")
	rflag := flag.Bool("R", false, "reset ranseed before each rx")
	flag.Parse()
	var efile *bufio.Scanner
	if len(flag.Args()) == 0 {
		efile = rx.MkScanner("-")
	} else if len(flag.Args()) == 1 {
		efile = rx.MkScanner(flag.Args()[0])
	} else {
		log.Fatal("usage: rxr [efile]")
	}
	rand.Seed(int64(time.Now().Nanosecond()))

	// load and process regexps
	for i := 0; efile.Scan(); i++ {
		fmt.Println()
		spec := efile.Text()
		if len(spec) > 0 && spec[0] == '#' { // if comment, not RE
			fmt.Println(spec)
			continue
		}
		fmt.Printf("regexp: %s\n", spec)
		t, e := rx.Parse(spec)
		if e != nil {
			fmt.Println("ERROR: ", e)
			continue
		}
		fmt.Printf("tree:   %v\n", t)
		fmt.Printf("length: %d to ", t.MinLen())
		x := t.MaxLen()
		if x >= 0 {
			fmt.Println(x)
		} else {
			fmt.Println("*")
		}

		augt := rx.Augment(t)    // make augmented tree
		dfa := rx.BuildDFA(augt) // make DFA from augmented tree

		if *rflag { // if -R given, reset random seed
			rand.Seed(0) // for independent, reproducible output
		}
		examples(dfa, t, 0) // gen and test w/ max repl of 0
		examples(dfa, t, 1) // ... and 1
		examples(dfa, t, 2) // ... and 2
		examples(dfa, t, 3) // ... and 3
		examples(dfa, t, 5) // ... and 5
		examples(dfa, t, 8) // ... and 8

		if !*qflag { // if -Q not given, print automata
			// show tree details
			treenodes(augt)
			// show followpos sets
			for _, m := range dfa.Leaves {
				fmt.Printf("p%d. %s => {", m.Posn, m)
				for d := range m.FollowPos {
					fmt.Print(" ", d.Posn)
				}
				fmt.Print(" }\n")

			}
			// dump DFA
			for _, d := range dfa.Dstates {
				showstate(dfa, d)
			}
		}
	}
	rx.CkErr(efile.Err())
}

const linemax = 79

//   Examples generates a line's worth of examples with max replication n.
func examples(dfa *rx.DFA, tree rx.Node, n int) {
	s := fmt.Sprintf("ex(%d):", n)
	nprinted := 0
	fmt.Print(s)
	ncolm := len(s)
	for {
		s := string(tree.Example(make([]byte, 0), n))
		t := rx.Protect(s)
		ncolm += 2 + len(t)
		if nprinted > 0 && ncolm > linemax {
			break
		}
		fmt.Printf("  %s", t)
		if !dfa.Accepts(s) {
			fmt.Print(" [FAIL]")
			ncolm += 7
		}
		nprinted++
	}
	fmt.Println()
}

//  Treenodes prints details of the parse tree.
func treenodes(tree rx.Node) {
	indent = ""
	tree.Walk(predump, postdump)
}

var indent string

func predump(d rx.Node) {
	indent = indent + "  "
	a := d.Data()
	fmt.Printf("%snode: {%t, ", indent[2:], a.Nullable)
	for k, _ := range a.FirstPos {
		fmt.Print(k)
	}
	fmt.Print(", ")
	for k, _ := range a.LastPos {
		fmt.Print(k)
	}
	fmt.Println("}", d)

}

func postdump(d rx.Node) {
	indent = indent[2:]
}

//  Showstate prints the contents of one DFA state.
func showstate(dfa *rx.DFA, d *rx.DFAstate) {

	// print index with "Accept" flag
	if dfa.AcceptBy(d) {
		fmt.Printf("s%d# {", d.Index)
	} else {
		fmt.Printf("s%d. {", d.Index)
	}

	// print position set
	for _, j := range d.Posns.Members() {
		fmt.Printf(" p%d", j)
	}
	fmt.Print(" }")

	// invert the transition map
	slist := &rx.BitSet{}
	xmap := make(map[int]*rx.BitSet)
	for c, d := range d.Dnext {
		j := d.Index
		v := xmap[j]
		if v == nil {
			v = &rx.BitSet{}
			xmap[j] = v
			slist.Set(uint(j))
		}
		v.Set(uint(c))
	}

	// now print by next-state with input symbols grouped
	for _, c := range slist.Members() {
		fmt.Printf(" %s:s%d", xmap[int(c)].Bracketed(), c)
	}
	fmt.Println()
}
