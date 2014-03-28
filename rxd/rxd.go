/*
	rxd.go -- regular expression to Dot

	usage:  rxd [-u] [-v] ["rexpr"...]

	Rxd builds a DFA that accepts one or more regular expressions
	and generates Dot (Graphviz) directives for displaying the graph.

	If -u is given, the unoptimized DFA is drawn.
	Otherwise, the minimized DFA is drawn.

	If -v is given, the output is written to a temporary file,
	drawn as a SVG, file, and the default viewer is invoked.
	For this to work, "dot" must be in the command path.

	If no expressions are given as command arguments, rxd reads
	expressions from standard input, one per line, and combines them.
	A line beginning with '#', or an empty line, is treated as a comment.

	Spring-2014 / gmt
*/
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"rx"
	"syscall"
)

func main() {

	uflag := flag.Bool("u", false, "draw unoptimized DFA")
	vflag := flag.Bool("v", false, "open in viewer")
	flag.Parse()

	// build a list of regular expressions
	elist := make([]string, 0)
	for i := 0; i < len(flag.Args()); i++ {
		elist = append(elist, flag.Args()[i])
	}
	if len(elist) == 0 {
		rx.LoadExpressions("-", func(x *rx.RegExParsed) {
			if x.IsExpr() {
				elist = append(elist, x.Expr)
			}
		})
	}

	// build a list of augmented parse trees
	tlist := make([]rx.Node, 0, len(elist))
	for _, e := range elist {
		ptree, err := rx.Parse(e)
		rx.CkErr(err)
		tlist = append(tlist, rx.Augment(ptree, uint(len(tlist))))
	}

	// build the DFA
	dfa := rx.MultiDFA(tlist)
	if !*uflag {
		dfa = dfa.Minimize()
	}

	// generate Dot output
	var label string
	if len(tlist) > 1 {
		label = fmt.Sprintf("(%d expressions)", len(tlist))
	} else {
		label = elist[0]
	}

	if *vflag {
		display(dfa, label)
	} else {
		dfa.ToDot(os.Stdout, label)
	}
}

//  display converts the dot output to SVG and starts a viewer.
func display(dfa *rx.DFA, label string) {
	syscall.Umask(077) // for temp files, override umask
	dotfile, err := ioutil.TempFile("", "rxd")
	rx.CkErr(err)
	dotname := dotfile.Name()
	dfa.ToDot(dotfile, label)
	dotfile.Close()
	svgname := dotname + ".svg"
	cmd := exec.Command("dot", "-Tsvg", dotname, "-o", svgname)
	err = cmd.Run()
	rx.CkErr(err)
	os.Remove(dotname)
	if runtime.GOOS == "darwin" { // if Macintosh
		err = exec.Command("open", "-W", svgname).Run()
	} else {
		err = exec.Command("xdg-open", svgname).Run()
	}
	rx.CkErr(err)
	// #%#% DISABLED: os.Remove(svgname)
	// We don't remove the temp file because we don't know when it's safe.
	// It's especially problematic when multiple views are open at once.
	// It would be nice to find a solution for this.
}
