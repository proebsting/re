
//  This code implements the beginnings of a web application
//  using Google App Engine
//  for inspecting and comparing regular expressions.

package webapp

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"rx"
)



var examples = []struct{ Expr, Caption string }{
	{`(0|1(01*0)*1)*`, "Binary number divisible by 3"},
	{`-?(0|[1-9]\d*)(\.\d+)?([eE][+-]?\d{1,3})?`, "JSON number"},
	{`\([2-9]\d\d\) [2-9]\d\d\-\d{4}`, "US telephone number"},
	{`([0-6]\d{2}|7([0-6]\d|7[012]))-\d{2}-\d{4}`, "US social security number"},
	{`(19|20)\d\d\-(0[1-9]|1[012])\-(0[1-9]|[12]\d|3[01])`, "ISO 8601 date"},
	{`([01]\d|2[0-3]):[0-5][0-9]:[0-5][0-9]Z`, "ISO 8601 time"},
	{`\w+@\w+(\.\w+)+`, "Naive e-mail address"},
}

//  examine presents a query page for examining a single expression
func examine(w http.ResponseWriter, r *http.Request) {
	putheader(w, r, "Inspection Query")
	putform(w, 1, "/details", "Enter a regular expression:")
	tExamples.Execute(w, examples)
	putfooter(w, r)
}

var tExamples = template.Must(template.New("examples").Parse(
	`<P>Or choose one of these examples:{{range .}}
<form action="/details" method="post"><div>
<input type="hidden" name="a0" value="{{.Expr}}">
<button class=link>{{.Caption}}</button></div></form>{{end}}
`))

//  details responds to an inspection request for a single expression
func details(w http.ResponseWriter, r *http.Request) {
	expr := r.FormValue("a0") // must read before any writing
	putheader(w, r, "Inspect Expression")
	tree, err := rx.Parse(expr)
	if err == nil {
		fmt.Fprintf(w, "<P>Regular Expression: %s\n", hx(expr))
		// must print (or at least stringize) tree before augmenting
		fmt.Fprintf(w, "<P>Initial Parse Tree: %s\n", hx(tree))

		//#%#% currently no protection against DOS by huge expr
		augt := rx.Augment(tree, 0)
		dfa := rx.BuildDFA(augt)
		dmin := dfa.Minimize()

		fmt.Fprintf(w, "<P>Augmented Tree: %s\n", hx(augt))
		fmt.Fprintf(w, "<h2>Examples</h2>\n<P>")
		genexamples(w, tree, 0)
		genexamples(w, tree, 1)
		genexamples(w, tree, 2)
		genexamples(w, tree, 3)
		genexamples(w, tree, 5)
		genexamples(w, tree, 8)

		nfaBuffer := &bytes.Buffer{}
		dfa.ShowNFA(nfaBuffer, "")
		fmt.Fprintf(w, "<h2>NFA</h2><PRE>\n%s</PRE>\n",
			hx(string(nfaBuffer.Bytes())))

		dfaBuffer := &bytes.Buffer{}
		dmin.DumpStates(dfaBuffer, "")
		fmt.Fprintf(w, "<h2>DFA</h2><PRE>\n%s</PRE>\n",
			hx(string(dfaBuffer.Bytes())))
	} else {
		showerror(w, err)
	}

	fmt.Fprint(w, "<h2>Try another?</h2>")
	putform(w, 1, "/details", "Enter a regular expression:")
	putfooter(w, r)
}

const linemax = 79 // max output line length for examples

//  genexamples writes a line of specimen strings matching the expression
func genexamples(w http.ResponseWriter, tree rx.Node, maxrepl int) {
	nprinted := 0
	ncolm := 0
	for {
		s := string(tree.Example(make([]byte, 0), maxrepl))
		t := rx.Protect(s)
		ncolm += 2 + len(t)
		if nprinted > 0 && ncolm > linemax {
			break
		}
		fmt.Fprintf(w, " %s &nbsp; ", hx(t))
		nprinted++
	}
	fmt.Fprint(w, "<BR>\n")
}
