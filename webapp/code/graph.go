// graph.go -- code for displaying DFA and NFA machines as lists or graphs

package webapp

import (
	"bytes"
	"fmt"
	"net/http"
	"rx"
)

//  multaut displays a multi-NFA and multi-DFA
func multaut(w http.ResponseWriter, r *http.Request) {

	// must read all input before writing anything
	exprlist := getexprs(r)
	nx := len(exprlist)

	// parse and echo the input
	treelist := make([]rx.Node, 0, nx)
	putheader(w, r, "Multi-expression Automata")
	fmt.Fprintf(w, "<P>%d expressions:\n", nx)
	for i, s := range exprlist {
		fmt.Fprintf(w, "<BR><B>e%d:</B> &nbsp; %s\n", i, hx(s))
		tree, err := rx.Parse(s)
		if !showerror(w, err) {
			treelist = append(treelist, rx.Augment(tree, i))
		}
	}

	if nx > 0 && len(treelist) == nx { // if no errors
		dfa := rx.MultiDFA(treelist) // build combined DFA
		dmin := dfa.Minimize()
		showaut(w, dmin, exprlist)
	}
	putfooter(w, r)
}

//  print NFA and DFA, with buttons linking to display page
func showaut(w http.ResponseWriter, dfa *rx.DFA, exprlist []string) {

	fmt.Fprintln(w, `<div><div class=lfloat>`)

	nfaBuffer := &bytes.Buffer{}
	dfa.ShowNFA(nfaBuffer, "")
	fmt.Fprintf(w, "<h2>NFA</h2><PRE>\n%s</PRE>\n",
		hx(string(nfaBuffer.Bytes())))
	askgraph(w, "/drawNFA", exprlist, "Draw the graph")

	fmt.Fprintln(w, `</div><div class=lstripe>`)
	dfaBuffer := &bytes.Buffer{}
	dfa.ShowStates(dfaBuffer, "")
	fmt.Fprintf(w, "<h2 class=noadvance>DFA</h2><PRE>\n%s</PRE>\n",
		hx(string(dfaBuffer.Bytes())))
	askgraph(w, "/drawDFA", exprlist, "Draw the graph")

	fmt.Fprintln(w, `</div></div><div class=reset></div>`)
}
