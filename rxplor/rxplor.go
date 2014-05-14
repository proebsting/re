/*
	rxplor.go - regular expression explorer

	The Swiss-army-knife utility for examining regular expressions.

	usage:  rxplor [options] [exprfile]

	Rxplor reads regular expressions from exprfile or from stdin
	and processes them either individually or all merged together.

	See func options for the numerous command options.  With no
	options, rxplor just lists the input expressions and any errors.

	spring 2014 / gmt
*/
package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"runtime/pprof"
	"rx"
	"rx/rsys"
	"strconv"
	"time"
	"unicode/utf8"
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
	if *opt['i'] { // if preceded by individual processing
		fmt.Println()
		rx.ShowLabel(os.Stdout, "MERGING EXPRESSIONS")
		fmt.Println()
	}

	dfa := rx.MultiDFA(trees)
	timestamp(fmt.Sprintf(
		"make merged DFA of %d states", len(dfa.Dstates)))
	dfa = showDFA(dfa, "Combined Tree", *opt['t'])

	// generate graphs if requested
	label := exprs[0].Expr
	if len(exprs) > 1 {
		label = fmt.Sprintf("(%d expressions)", len(exprs))
	}
	if *val['D'] != "" {
		rx.WriteGraph(*val['D'], func(w io.Writer) {
			dfa.ToDot(w, label)
		})
	}
}

// load slurps up expressions, returning list and augmented parse tree list
func load() ([]*rx.RegExParsed, []rx.Node) {
	exprs := make([]*rx.RegExParsed, 0) // expression structure list
	trees := make([]rx.Node, 0)         // augmented parse tree list

	rx.LoadFromScanner(input, func(l *rx.RegExParsed) {
		echo(l, len(exprs))
		if l.Tree != nil { // if a real expression
			augtree := rx.Augment(l.Tree, len(trees))
			trees = append(trees, augtree)
			exprs = append(exprs, l)
			if *opt['i'] {
				individual(l, len(exprs))
			}
		}
	})
	babble("%d expression(s) loaded\n", rx.InputRegExCount)
	if rx.InputErrorCount != 0 && (!errsilent || verbose) {
		fmt.Printf("(%d expression(s) rejected)\n",
			rx.InputErrorCount)
	}
	return exprs, trees
}

// echo prints an input record and errors depending on command options
func echo(l *rx.RegExParsed, i int) {
	if !l.IsExpr() && l.Meta != nil { // if metadata line
		return // will print this later; it's been saved
	}
	if l.Err != nil { // if an error
		if !errsilent { // print error unless -y
			fmt.Printf("\nERROR:   %s\n    %s\n", l.Expr, l.Err)
			if *opt['x'] { // if immediate exit requested
				log.Fatal(l.Err)
			}
		}
		return
	}
	if l.Tree == nil { // if a comment
		if listopt { // print if wanted
			fmt.Printf("\n         %s\n", l.Expr)
		}
		return
	}

	// must be a real expression
	if *opt['i'] || listopt {
		fmt.Println()
		if listopt { // if to echo metadata
			// print accumulated metadata
			for _, k := range rx.KeyList(l.Meta) {
				fmt.Printf("         #%s: %v\n", k, l.Meta[k])
			}
		}
		fmt.Printf("expr %d: %s\n", i, l.Expr)
	}
}

// individual handles separate processing of each input regex
func individual(l *rx.RegExParsed, i int) {
	babble("tree:   %v\n", l.Tree)
	augt := rx.Augment(l.Tree, i)
	babble("augmnt: %v\n", augt)
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
	showDFA(dfa, "Annotated Tree", false)
}

//   examples generates a line's worth of examples with max replication n.
//   Each example is also tested against the DFA.
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

// showDFA performs common actions for individual and merged DFAs
// it returns the ultimate (original or minimized) dfa
func showDFA(dfa *rx.DFA, treelabel string, showtime bool) *rx.DFA {
	if *opt['p'] {
		dfa.ShowTree(os.Stdout, dfa.Tree, treelabel)
	}
	if *opt['n'] {
		dfa.ShowNFA(os.Stdout, "NFA")
	}
	if *opt['d'] {
		dfa.ShowStates(os.Stdout, "Unoptimized DFA")
	}
	if !*opt['u'] {
		if showtime {
			rsys.Interval() // reset for timing
		}
		dfa = dfa.Minimize()
		if showtime {
			timestamp(fmt.Sprintf(
				"minimize to %d states", len(dfa.Dstates)))
		}
		if *opt['d'] {
			dfa.ShowStates(os.Stdout, "Minimized DFA")
		}
	}
	if *opt['h'] {
		synthx(dfa)
	}
	return dfa
}

//  synthx generates and prints synthetic examples from a DFA.
func synthx(dfa *rx.DFA) {
	_, isMulti := dfa.Tree.(*rx.AltNode) // true if MultiDFA
	synthx := dfa.Synthesize()           // synthesize examples
	rx.ShowLabel(os.Stdout, "Examples from DFA")
	for _, x := range synthx {
		fmt.Printf("s%d:  %s", x.State, rx.Protect(x.Example))
		if isMulti {
			fmt.Printf("  %s\n", x.RXset)
		} else {
			fmt.Println()
		}
	}
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

	fo('l', "list input comments and metadata")
	fo('p', "show parse tree nodes with nullable, first, last")
	fo('n', "show positions and followsets of NFA")
	fo('d', "show states and transitions of DFA")
	fo('g', "show synthetic examples of matching strings from parse")
	fo('h', "show examples with reference to DFA states")
	fo('t', "output timing measurements")
	fo('v', "enable running commentary")
	fo('a', "enable all output options") // all of the above
	fo('x', "exit immediately on erroneous input")
	fo('y', "silently ignore errors")
	fo('z', "enable debug tracing")

	vo('I', "initial seed for randomization")
	// vo('T', "input file of candidates accept/reject grid")

	// file format for N and D depend on extension of filename supplied
	// a filename of - generates a temporary SVG file and opens a viewer
	// vo('N', "output file for NFA graph (*.dot/.gif/.pdf/.png/.svg/-)")
	vo('D', "output file for DFA graph (*.dot/.gif/.pdf/.png/.svg/-)")

	// vo('X', "output file for DFA-based examples (JSON)")

	vo('P', "output file for profiling data (binary)")

	flag.Parse() // parse command args to set values

	imply("a", "lpndghtv") // -a implies several others
	imply("NDXT", "m")     // any of {-N -D -X -T} implies -m

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
