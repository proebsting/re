/*
	rxq.go -- regular expression query

	usage:  rxr "rexpr" [file]

	Rxq reads strings, one per line, from file (default stdin).
	Each string is tested against the regular expression rexpr.

	Spring-2014 / gmt
*/
package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"rx"
)

func main() {
	var ifile *bufio.Scanner
	if len(os.Args) == 2 {
		ifile = rx.MkScanner("-")
	} else if len(os.Args) == 3 {
		ifile = rx.MkScanner(os.Args[2])
	} else {
		log.Fatal("usage: rxq \"rexpr\" [file]")
	}
	spec := os.Args[1]
	fmt.Printf("regexp: %s\n", spec)
	t, e := rx.Parse(spec)
	if e != nil {
		log.Fatal(e)
	}
	dfa, _ := rx.BuildDFA(t)

	// load and process candidate strings
	for i := 0; ifile.Scan(); i++ {
		s := ifile.Text()
		if dfa.Accepts(s) {
			fmt.Println("accept: ", s)
		} else {
			fmt.Println("REJECT: ", s)
		}
	}
	rx.CkErr(ifile.Err())
}
