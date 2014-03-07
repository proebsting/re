/*
	rxd.go -- regular expression to Dot

	usage:  rxd "rexpr"

	Rxd builds a DFA that accepts a given regular expression
	and generates Dot-format directives for displaying the graph.

	Spring-2014 / gmt
*/
package main

import (
	"fmt"
	"log"
	"os"
	"rx"
	"strconv"
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

	fmt.Println("//", spec)
	fmt.Println("digraph DFA {")
	fmt.Printf("label=%s\n", strconv.Quote(spec))
	fmt.Println("node [shape=circle, height=.3, margin=0, fontsize=10]")
	fmt.Println("s0 [shape=triangle, regular=true]")
	for _, src := range dfa.Dstates {
		if dfa.AcceptBy(src) {
			fmt.Printf("s%d[shape=doublecircle]\n", src.Index)
		}
		slist, xmap := dfa.InvertMap(src)
		for _, dst := range slist.Members() {
			fmt.Printf("s%d->s%d[label=\"%s\"]\n",
				src.Index, dst, xmap[int(dst)].Bracketed())
		}
	}
	fmt.Println("}")
}
