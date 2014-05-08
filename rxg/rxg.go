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
	"flag"
	"fmt"
	"math/rand"
	"os"
	"rx"
	"time"
)

type RegEx struct { // one regexp for JSON output
	Index int    // index number
	Rexpr string // regular expression
}

type Example struct { // one example for JSON output
	State   int    // index of accepting state in DFA
	RXset   []int  // set of matching regular expression indexes
	Example string // example string
}

func main() {

	rflag := flag.Bool("R", false, "reproducible output")
	flag.Parse()
	if *rflag {
		rand.Seed(0)
	} else {
		rand.Seed(int64(time.Now().Nanosecond()))
	}

	// load and process regexps
	exprs := make([]*RegEx, 0)
	tlist := make([]rx.Node, 0)
	rx.LoadExpressions(rx.OneInputFile(), func(l *rx.RegExParsed) {
		if l.Err != nil {
			fmt.Fprintln(os.Stderr, l.Err)
		}
		if l.Tree != nil {
			atree := rx.Augment(l.Tree, len(tlist))
			tlist = append(tlist, atree)
			exprs = append(exprs, &RegEx{len(exprs), l.Expr})
		}
	})

	// echo the input with index numbers
	fmt.Print(`{"Expressions":`)
	rx.Jlist(os.Stdout, exprs)
	fmt.Println(",")

	// build the DFA and produce examples
	synthx := rx.MultiDFA(tlist).Synthesize()

	// convert into expected form with int array replacing BitSet
	results := make([]*Example, 0, len(synthx))
	for _, x := range synthx {
		results = append(results,
			&Example{x.State, x.RXset.Members(), x.Example})
	}

	// output the array of synthesized examples
	fmt.Print(`"Examples":`)
	rx.Jlist(os.Stdout, results)
	fmt.Println("}")
}
