/*
	rxr.go -- regular expression reader

	usage:  rxr [-Z] [exprfile]

	-Z	reset random seed for each regexp (for testing consistency)

	Rxr reads regular expressions, one per line, from efile,
	and does something with them (details constantly changing).
	A line beginning with '#' is treated as a comment.
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
		if len(spec) > 0 && spec[0] == '#' {	// if comment, not RE
			fmt.Println(spec)
			continue
		}
		fmt.Printf("regexp: %s\n", spec)
		t := rx.Parse(spec)
		if t == nil {
			fmt.Println("(ERROR)")
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
		if (*zflag) {		// if -Z given, reset random seed
			rand.Seed(0)	// for independent, reproducible output
		}
		examples(t, 0)		// gen examples with max repl of 0
		examples(t, 1)		// ... and 1
		examples(t, 2)		// ... and 2
		examples(t, 3)		// ... and 3
		examples(t, 5)		// ... and 5
		examples(t, 8)		// ... and 8
	}
	rx.CkErr(efile.Err())
}

//  generate a line's worth of examples from a rx
const linemax = 79
func examples(rx rx.Node, n int) {
	s := fmt.Sprintf("ex(%d):  %s", n, rx.Example(make([]byte, 0), n))
	cc := len(s)
	fmt.Print(s)
	for {
		t := rx.Example(make([]byte, 0), n)
		cc += 2 + len(t)
		if cc > linemax {
			break
		}
		fmt.Printf("  %s", t)
	}
	fmt.Println()
}
