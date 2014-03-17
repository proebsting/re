/*
	rxd.go -- regular expression to Dot

	usage:  rxd [-v] ["rexpr"...]

	Rxd builds a DFA that accepts one or more regular expressions
	and generates Dot (Graphviz) directives for displaying the graph.

	If -v is given, the output is written to a temporary file, drawn
	as a GIF file, and the default viewer is invoked.  For this to work,
	"dot" must be in the command path.

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

	vflag := flag.Bool("v", false, "view PDF graph")
	flag.Parse()

	// build a list of regular espressions
	elist := make([]string, 0)
	for i := 0; i < len(flag.Args()); i++ {
		elist = append(elist, flag.Args()[i])
	}
	if len(elist) == 0 {
		ifile := rx.MkScanner("-")
		for ifile.Scan() {
			e := ifile.Text()
			if !rx.IsComment(e) {
				elist = append(elist, ifile.Text())
			}
		}
		rx.CkErr(ifile.Err())
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

// display on screen by converting dot to PDF and invoking a viewer
func display(dfa *rx.DFA, label string) {
	syscall.Umask(077) // for temp files, override umask
	dotfile, err := ioutil.TempFile("", "rxd")
	rx.CkErr(err)
	dotname := dotfile.Name()
	dfa.ToDot(dotfile, label)
	dotfile.Close()
	gifname := dotname + ".gif"
	cmd := exec.Command("dot", "-Tgif", dotname, "-o", gifname)
	err = cmd.Run()
	rx.CkErr(err)
	os.Remove(dotname)
	if runtime.GOOS == "darwin" { // if Macintosh
		err = exec.Command("open", "-W", gifname).Run()
	} else {
		err = exec.Command("xdg-open", gifname).Run()
	}
	rx.CkErr(err)
	os.Remove(gifname)
}
