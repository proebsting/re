/*
	rxg.go -- regular expression multiple example generator

	usage:  rxg [-R] [exprfile]

	Rxg reads a set of regular expressions and synthesizes one example
	corresponding to each accepting state in the combined DFA.

	-R	produce reproducible output by using a fixed random seed

	Input is one unadorned regular expression per line.
	A line beginning with '#', or an empty line, is treated as a comment.

	Output is a struct of two arrays in JSON format.  The first array
	lists the regular expressions with input numbers. The second lists
	examples with state numbers and sets of matching regular expressions.

	Example:
	For the input
		\d+
		\d*[1-9]
		[1-9]\d*
	the output is:
		{"Expressions":[
		{"Index":0,"Rexpr":"\\d+"},
		{"Index":1,"Rexpr":"\\d*[1-9]"},
		{"Index":2,"Rexpr":"[1-9]\\d*"}
		],
		"Examples":[
		{"State":1,"RXset":[0],"Example":"0"},
		{"State":2,"RXset":[0,1,2],"Example":"7"},
		{"State":3,"RXset":[0,1],"Example":"02"},
		{"State":4,"RXset":[0,2],"Example":"70"}
		]}

	Spring-2014 / gmt
*/
package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"rx"
	"time"
)

type RegEx struct { // one rx for JSON output
	Index int    // index number
	Rexpr string // regular expression
}

func main() {

	rflag := flag.Bool("R", false, "reproducible output")
	flag.Parse()

	var efile *bufio.Scanner
	if len(flag.Args()) == 0 {
		efile = rx.MkScanner("-")
	} else if len(flag.Args()) == 1 {
		efile = rx.MkScanner(flag.Args()[0])
	} else {
		log.Fatal("usage: rxg [-R] [exprfile]")
	}
	if *rflag {
		rand.Seed(0)
	} else {
		rand.Seed(int64(time.Now().Nanosecond()))
	}

	// load and process regexps
	exprs := make([]RegEx, 0)
	tlist := make([]rx.Node, 0)
	for efile.Scan() {
		line := efile.Text()
		if rx.IsComment(line) {
			continue
		}
		exprs = append(exprs, RegEx{len(exprs), line})
		ptree, err := rx.Parse(line)
		rx.CkErr(err)
		tlist = append(tlist, rx.Augment(ptree, uint(len(tlist))))
	}
	rx.CkErr(efile.Err())

	// echo the input with index numbers
	fmt.Print(`{"Expressions":`)
	rx.Jlist(os.Stdout, exprs)
	fmt.Println(",")

	// build the DFA and produce examples
	results := rx.MultiDFA(tlist).Synthesize()

	// output the array of synthesized examples
	fmt.Print(`"Examples":`)
	rx.Jlist(os.Stdout, results)
	fmt.Println("}")
}
