/*
	rxr.go -- regular expression reader

	usage:  rxr [-A] [-Z] [exprfile]

	-A	disable printing of automata details
	-Z	reset random seed for each regexp (for testing consistency)

	Rxr reads regular expressions, one per line, from efile.
	A line beginning with '#' is treated as a comment.

	Rxr currently echoes each expression, shows the parse tree,
	prints some statistics, and generates some examples.
	This will certainly be evolving over time.

	The output is indented to be informative, and unprintables are escaped,
	but will not always be unambiguous if pattern metacharacters are used
	as ordinary matching characters

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
	aflag := flag.Bool("A", false, "disable automata printing")
	zflag := flag.Bool("Z", false, "reset ranseed before each rx")
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
		if *zflag { // if -Z given, reset random seed
			rand.Seed(0) // for independent, reproducible output
		}
		examples(t, 0) // gen examples with max repl of 0
		examples(t, 1) // ... and 1
		examples(t, 2) // ... and 2
		examples(t, 3) // ... and 3
		examples(t, 5) // ... and 5
		examples(t, 8) // ... and 8
		if !*aflag {   // if -A not given, print automata
			bigdump(t)
		}
	}
	rx.CkErr(efile.Err())
}

const linemax = 79

//   Examples generates a line's worth of examples with max replication n.
func examples(x rx.Node, n int) {
	s := fmt.Sprintf("ex(%d):  %s",
		n, rx.Protect(string(x.Example(make([]byte, 0), n))))
	cc := len(s)
	fmt.Print(s)
	for {
		s = rx.Protect(string(x.Example(make([]byte, 0), n)))
		cc += 2 + len(s)
		if cc > linemax {
			break
		}
		fmt.Printf("  %s", s)
	}
	fmt.Println()
}

//  Bigdump prints details of the parse tree.
func bigdump(x rx.Node) {
	indent := ""
	x.Walk(&indent, previsit, postvisit)
}

func previsit(d rx.Node, v interface{}) {
	indent := v.(*string)
	*indent = *indent + "  "
	a := d.Data()
	fmt.Printf("%snode: {%t, ", (*indent)[2:], a.Nullable)
	for k, _ := range a.FirstPos {
		fmt.Print(k)
	}
	fmt.Print(", ")
	for k, _ := range a.LastPos {
		fmt.Print(k)
	}
	fmt.Println("}", d)

}

func postvisit(d rx.Node, v interface{}) {
	indent := v.(*string)
	*indent = (*indent)[2:]
}
