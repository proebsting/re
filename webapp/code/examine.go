//  examine.go -- code for inspecting a single expression in detail

package webapp

import (
	"fmt"
	"net/http"
	"rx"
	"strings"
)

const linemax = 79 // max output line length for generated examples

//  examine presents a query page for examining a single expression
func examine(w http.ResponseWriter, r *http.Request) {
	putheader(w, r, "Inspection Query")
	fmt.Fprintln(w, `<P> Here you can specify
a regular expression to generate several kinds of data:
parse trees, synthetic examples, and the automata state lists.
Links from the results page produce diagrams of either the
<A HREF="http://en.wikipedia.org/wiki/Nondeterministic_finite_automaton">NFA</A>
or
<A HREF="http://en.wikipedia.org/wiki/Deterministic_finite_automaton">DFA</A>
for the language.
`)
	putform(w, "/details", "Enter a regular expression:", 1, nil, 0, nil)
	fmt.Fprintln(w, `<P>Or choose one of these examples:`)
	for _, x := range examples {
		genexam(w, x.Caption, x.Expr)
	}
	putfooter(w, r)
}

var examples = []struct{ Expr, Caption string }{
	{HomeExample, HomeExLabel},
	{`-?(0|[1-9]\d*)(\.\d+)?([eE][+-]?\d+)?`, "JSON number"},
	{`(0|1(01*0)*1)*`, "Binary number divisible by 3"},
	{`A(BA)*B?|B(AB)*A?|A?(BA)*B?C((A(BA)*B?|B(AB)*A?)C)*(A(BA)*B?|B(AB)*A?)?`,
		"ABCs with no letter doubled"},
	{`\([2-9]\d\d\) [2-9]\d\d\-\d{4}`, "US telephone number"},
	{`[0-8]\d\d-\d\d-\d{4}`, "US social security number"},
	{`(19|20)\d\d\-(0[1-9]|1[012])\-(0[1-9]|[12]\d|3[01])`, "ISO 8601 date"},
	{`([01]\d|2[0-3]):[0-5]\d:[0-5]\dZ`, "ISO 8601 time"},
	{`\w+@\w+(\.\w+)+`, "Naive e-mail address"},
}

//  details responds to an inspection request for a single expression
func details(w http.ResponseWriter, r *http.Request) {
	expr := r.FormValue("x0")      // must read before any writing
	expr = strings.TrimSpace(expr) // trim leading/trailing blanks
	putheader(w, r, "Inspect Expression")
	tree, err := rx.Parse(expr)
	if err == nil {
		fmt.Fprintf(w, "<P>Regular Expression: %s\n", hx(expr))
		// must print (or at least stringize) tree before augmenting
		fmt.Fprintf(w, "<P>Initial Parse Tree: %s\n", hx(tree))

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

		showaut(w, dmin, []string{expr})
	} else {
		fmt.Fprint(w, "<P>")
		showerror(w, err)
	}

	fmt.Fprint(w, "<h2>Try another?</h2>")
	putform(w, "/details", "Enter a regular expression:",
		1, []string{expr}, 0, nil)
	putfooter(w, r)
}

//  genexamples writes a line of specimen strings matching the expression
func genexamples(w http.ResponseWriter, tree rx.Node, maxrepl int) {
	nprinted := 0
	ncolm := 0
	for {
		s := rx.Specimen(tree, maxrepl)
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
