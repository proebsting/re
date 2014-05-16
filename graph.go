//  graph.go -- graphical output via Dot (GraphViz)

package rx

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

//  DFA.GraphNFA generates a Dot (GraphViz) representation of the NFA.
func (dfa *DFA) GraphNFA(f io.Writer, label string) {
	fmt.Fprintln(f, "//", label)
	fmt.Fprintln(f, "digraph NFA {")
	fmt.Fprintf(f, "label=%s\n", strconv.Quote(label))
	fmt.Fprintln(f, "node [shape=circle, height=.3, margin=0, fontsize=10]")
	startshape := "triangle"
	for _, p := range dfa.Tree.Data().FirstPos.Members() {
		if IsAccept(dfa.Leaves[p]) {
			startshape = "doublecircle"
		} else {
			fmt.Fprintf(f, "i->p%d[label=\" %s\"]\n",
				p, dfa.Leaves[p])
		}
	}
	fmt.Fprintf(f, "i [shape=%s, regular=true, label=\"\"]\n", startshape)
	for i, l := range dfa.Leaves {
		if IsAccept(l) {
			continue
		}
		fmt.Fprintf(f, "p%d [label=\"p%d\"]\n", i, i)
		for _, p := range l.FollowPos.Members() {
			if IsAccept(dfa.Leaves[p]) {
				fmt.Fprintf(f, "p%d [shape=doublecircle]\n", i)
			} else {
				fmt.Fprintf(f, "p%d->p%d[label=\" %s\"]\n",
					i, p, dfa.Leaves[p])
			}
		}
	}
	fmt.Fprintln(f, "}")
}

//  DFA.ToDot generates a Dot (GraphViz) representation of the DFA.
func (dfa *DFA) ToDot(f io.Writer, label string) {
	fmt.Fprintln(f, "//", label)
	fmt.Fprintln(f, "digraph DFA {")
	fmt.Fprintf(f, "label=%s\n", strconv.Quote(label))
	fmt.Fprintln(f, "node [shape=circle, height=.3, margin=0, fontsize=10]")
	fmt.Fprintln(f, "s0 [shape=triangle, regular=true]")
	for _, src := range dfa.Dstates {
		if src.AccSet != nil {
			fmt.Fprintf(f, "s%d[shape=doublecircle]\n", src.Index)
		}
		slist, xmap := src.InvertMap()
		for _, dst := range slist.Members() {
			fmt.Fprintf(f, "s%d->s%d[label=\"%s\"]\n",
				src.Index, dst, xmap[dst].Bracketed())
		}
	}
	fmt.Fprintln(f, "}")
}

//  WriteGraph writes a graph based on the extension of the given filename.
//	*.dot -> Dot (GraphViz) format  (default for unrecognized forms)
//	*.gif -> GIF (Graphics Interchange Format)
//	*.pdf -> PDF (Portable Document Format)
//	*.png -> PNG (Portable Network Graphics)
//	*.svg -> SVG (Scalable Vector Graphics)
//
//  The argument genfunc is a function to actually generate the Dot output.
//  If another format is wanted, output is written to a temporary file and
//  then "dot" is run from the path to convert it.
//
//  If the filename is "-", another temporary file is written in SVG format
//  and a viewer is opened.  This temporary file is never deleted because we
//  don't know when it's safe to remove it.
func WriteGraph(filename string, genfunc func(io.Writer)) {
	var err error
	var otype string     // output conversion type
	var dotfile *os.File // output file for Dot format

	// check what type of output is wanted
	switch {
	case filename == "-":
		otype = "-Tsvg"
	case strings.HasSuffix(filename, ".gif"):
		otype = "-Tgif"
	case strings.HasSuffix(filename, ".pdf"):
		otype = "-Tpdf"
	case strings.HasSuffix(filename, ".png"):
		otype = "-Tpng"
	case strings.HasSuffix(filename, ".svg"):
		otype = "-Tsvg"
	default: // must want .dot format as ultimate output
		dotfile, err = os.Create(filename)
		CkErr(err)
	}
	if dotfile == nil { // if we need to use a temporary file
		dotfile, err = ioutil.TempFile("", "rxplor")
		CkErr(err)
	}

	// generate the Dot file
	genfunc(dotfile)
	CkErr(dotfile.Close())
	if otype == "" { // if nothing more to do
		return
	}

	// convert from Dot format to desired output format
	dotname := dotfile.Name()
	outname := filename
	if outname == "-" {
		outname = dotname + ".svg"
	}
	CkErr(exec.Command("dot", otype, dotname, "-o", outname).Run())
	os.Remove(dotname)
	if filename != "-" {
		return // no viewer wanted
	}

	// run a viewer
	if runtime.GOOS == "darwin" { // if Macintosh
		CkErr(exec.Command("open", "-W", outname).Run())
	} else {
		CkErr(exec.Command("xdg-open", outname).Run())
	}
	// DISABLED: os.Remove(outname)
	// We don't remove the temp file because we don't know when it's safe.
	// It's especially problematic when multiple views are open at once.
	// It would be nice to find a solution for this.
}
