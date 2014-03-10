/*
	rxd.go -- regular expression to Dot

	usage:  rxd "rexpr"

	Rxd builds a DFA that accepts a given regular expression
	and generates Dot-format directives for displaying the graph.

	Spring-2014 / gmt
*/
package main

import (
	"log"
	"os"
	"rx"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatal("usage: rxd \"rexpr\"")
	}
	spec := os.Args[1]
	dfa, err := rx.Compile(spec)
	if err != nil {
		log.Fatal(err)
	}
	dfa.ToDot(os.Stdout, spec)
}
