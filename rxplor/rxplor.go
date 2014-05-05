/*
	rxplor.go - regular expression explorer

	#%#% A WORK IN PROGRESS -- NOT YET FULLY IMPLEMENTED
	#%#% This is the experimental Swiss-army-knife approach

	usage:  rxplor [options] [exprfile]

	Rxplor reads regular expressions from exprfile or from stdin
	and processes them either individually or all merged together.

	See func options for the command options.

	spring 2014 / gmt
*/
package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"runtime/pprof"
	"rx"
	"rx/rsys"
	"strconv"
	"time"
)

// command line options and values
var opt = make(map[rune]*bool)   // values of flag options
var val = make(map[rune]*string) // values of string options

var listopt bool   // list input option
var verbose bool   // global verbosity flag
var errsilent bool // ignore errors silently

var input *bufio.Scanner // input file

const linemax = 79 // output line size max for examples

// main is the overall controller
func main() {
	setup()                      // initialize
	defer pprof.StopCPUProfile() // may have started profiling

	exprs, trees := load() // load input
	timestamp(fmt.Sprintf(
		"load %d expressions", len(exprs)))

	if !*opt['m'] { // if nothing uses a combined DFA
		return
	}

	dfa := rx.MultiDFA(trees)
	if *opt['p'] {
		dfa.ShowTree(os.Stdout, dfa.Tree, "Combined Tree")
	}
	if *opt['n'] {
		dfa.ShowNFA(os.Stdout, "NFA")
	}
	if *opt['d'] {
		dfa.DumpStates(os.Stdout, "Unoptimized DFA")
	}
	timestamp(fmt.Sprintf(
		"make merged DFA of %d states", len(dfa.Dstates)))

	if !*opt['u'] {
		mini := dfa.Minimize()
		if *opt['d'] {
			dfa.DumpStates(os.Stdout, "Minimized DFA")
		}
		timestamp(fmt.Sprintf(
			"minimize to %d states", len(mini.Dstates)))
	}
}

// load slurps up expressions, returning list and augmented parse tree list
func load() ([]*rx.RegExParsed, []rx.Node) {
	exprs := make([]*rx.RegExParsed, 0) // expression structure list
	trees := make([]rx.Node, 0)         // augmented parse tree list

	rx.LoadFromScanner(input, func(l *rx.RegExParsed) {
		echo(l, len(exprs))
		if l.Tree != nil { // if a real expression
			augtree := rx.Augment(l.Tree, uint(len(trees)))
			trees = append(trees, augtree)
			exprs = append(exprs, l)
			if *opt['i'] {
				individual(l, len(exprs))
			}
		}
	})
	babble("%d expression(s) loaded\n", rx.InputRegExCount)
	if rx.InputErrorCount != 0 {
		fmt.Printf("(%d expression(s) rejected)\n",
			rx.InputErrorCount)
	}
	return exprs, trees
}

// echo prints an input record and errors depending on command options
func echo(l *rx.RegExParsed, i int) {
	if !l.IsExpr() && l.Meta != nil { // if metadata line
		return // will print this later
	}
	if listopt { // if need to echo
		fmt.Println()
		// print accumulated metadata
		for _, k := range rx.KeyList(l.Meta) {
			fmt.Printf("         #%s: %v\n", k, l.Meta[k])
		}
	}
	if l.Err != nil && !errsilent { // print error unless -y
		fmt.Printf("ERROR:   %s\n    %s\n", l.Expr, l.Err)
		if *opt['x'] { // if immediate exit requested
			log.Fatal(l.Err)
		}
	} else if listopt { // print any input with -l
		if l.Tree != nil { // if an expression, show ordinal
			fmt.Printf("expr %d: %s\n", i, l.Expr)
		} else { // otherwise print as comment
			fmt.Printf("         %s\n", l.Expr)
		}
	}
}

// individual handles separate processing of each input regex
func individual(l *rx.RegExParsed, i int) {
	if listopt {
		babble("tree:   %v\n", l.Tree)
	}
	//#%#% gen examples here?
	augt := rx.Augment(l.Tree, uint(i))
	if listopt {
		babble("augmnt: %v\n", augt)
	}
	dfa := rx.BuildDFA(augt)

	if *opt['g'] {
		rx.ShowLabel(os.Stdout, "Examples")
		examples(dfa, l.Tree, 0) // gen and test w/ max repl of 0
		examples(dfa, l.Tree, 1) // ... and 1
		examples(dfa, l.Tree, 2) // ... and 2
		examples(dfa, l.Tree, 3) // ... and 3
		examples(dfa, l.Tree, 5) // ... and 5
		examples(dfa, l.Tree, 8) // ... and 8
	}
	if *opt['p'] {
		dfa.ShowTree(os.Stdout, augt, "Annotated Tree")
	}
	if *opt['n'] {
		dfa.ShowNFA(os.Stdout, "NFA")
	}
	if *opt['d'] {
		dfa.DumpStates(os.Stdout, "Unoptimized DFA")
	}
	if !*opt['u'] {
		dfa = dfa.Minimize()
		if *opt['d'] {
			dfa.DumpStates(os.Stdout, "Minimized DFA")
		}
	}
}

//   examples generates a line's worth of examples with max replication n.
//   Each example is also tested against the DFA.
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
		if dfa.Accepts(s) == nil {
			fmt.Print(" [FAIL]")
			ncolm += 7
		}
		nprinted++
	}
	fmt.Println()
}

// setup performs global initialization actions
func setup() {
	rsys.Interval() // initialize timestamps
	options()       // process command-line options
	if verbose {
		showopts() // show command options
	}

	// handle -P (profiling) option
	pfname := *val['P'] // profiling filename
	if pfname != "" {   // if set
		pfile, err := os.Create(pfname)
		rx.CkErr(err)
		pprof.StartCPUProfile(pfile)
	}

	// set up input scanning
	if *val['e'] != "" { // if -e specified
		if len(flag.Args()) > 0 {
			log.Fatal("-e option precludes reading " + flag.Arg(0))
		}
		input = bufio.NewScanner(bytes.NewReader([]byte(*val['e'])))
	} else {
		input = rx.MkScanner(rx.OneInputFile())
	}

	timestamp("init")
}

//  options defines the command line options to be accepted.
func options() {

	// $%$% TODO:
	// commented-out options are not (yet?) implemented
	// (remember to add them back to "imply" call when they're ready)
	// add option to generate clusters? to time combinations?

	vo('e', "single expression value (instead of reading input)")
	fo('i', "treat input expressions individually")
	fo('m', "merge all input expressions into one large DFA")
	fo('u', "don't minimize (optimize) any DFA produced")

	fo('l', "list input as read")
	fo('p', "show parse tree nodes with nullable, first, last")
	fo('n', "show positions and followsets of NFA")
	fo('d', "show states and transitions of DFA")
	fo('g', "show synthetic examples of matching strings from parse")
	// fo('h', "show examples with reference to DFA states")
	fo('t', "output timing measurements")
	fo('v', "enable running commentary")
	fo('a', "enable all output options") // all of the above
	// fo('q', "suppress default output")
	fo('x', "exit immediately on erroneous input")
	fo('y', "silently ignore errors")
	fo('z', "enable debug tracing")

	vo('I', "initial seed for randomization")
	// vo('T', "input file of candidates accept/reject grid")

	// file format for N and D depend on extension of filename supplied
	// a filename of - generates a temporary SVG file and opens a viewer
	// vo('N', "output file for NFA graph (.dot, .gif, .pdf, .svg, -)")
	// vo('D', "output file for DFA graph (.dot, .gif, .pdf, .svg, -)")

	// vo('X', "output file for DFA-based examples (JSON)")

	vo('P', "output file for profiling data (binary)")

	flag.Parse() // parse command args to set values

	imply("a", "lpndgtv") // -a implies several others
	imply("NDXT", "m")    // any of {-N -D -X -T} implies -m
	if !*opt['m'] {
		*opt['i'] = true // if not -m, then must have -i
	}
	if *val['I'] == "" {
		s := fmt.Sprintf("%d", 1+time.Now().Nanosecond()%9998)
		val['I'] = &s
	}

	// set globals based on options
	listopt = *opt['l']
	verbose = *opt['v']
	errsilent = *opt['y']
	rx.DBG_MIN = *opt['z'] && *opt['d']

	i, err := strconv.Atoi(*val['I'])
	rx.CkErr(err)
	rand.Seed(int64(i))
}

//  showopts prints the options as set
func showopts() {
	fmt.Print("Options:")
	for ch := 'A'; ch <= 'z'; ch++ {
		if opt[ch] != nil && *opt[ch] {
			fmt.Printf(" -%c", ch)
		}
	}
	for ch := 'A'; ch <= 'z'; ch++ {
		if val[ch] != nil && *val[ch] != "" {
			fmt.Printf(" -%c %s", ch, *val[ch])
		}
	}
	fmt.Println()
}

//  fo defines a flag (boolean) option
func fo(ch rune, label string) {
	opt[ch] = flag.Bool(string(ch), false, label)
}

//  vo defines a value (string) option
func vo(ch rune, label string) {
	val[ch] = flag.String(string(ch), "", label)
}

//  imply(a,b) selects every option in b if any in a was specified
func imply(ifseen string, implies string) {
	seen := false
	for _, ch := range ifseen {
		pb := opt[ch]
		pv := val[ch]
		seen = seen || (pb != nil && *pb)
		seen = seen || (pv != nil && *pv != "")
	}
	if seen {
		for _, ch := range implies {
			*opt[ch] = true
		}
	}
}

// babble prints its args via Printf only if the global "verbose" is set
func babble(f string, args ...interface{}) {
	if verbose {
		fmt.Printf(f, args...)
	}
}

//  timestamp issues a time-stamped comment line, if enabled by -t
func timestamp(s string) {
	if *opt['t'] {
		rsys.ShowInterval(s)
	}
}
