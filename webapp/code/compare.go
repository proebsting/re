// compare.go -- code for comparing multiple regular expressions

package webapp

import (
	"fmt"
	"html/template"
	"net/http"
	"rx"
)

var DRAWLINE = "\x7F\x7F" // special flag for separator in grid

//  compare presents a page asking for multiple expressions
func compare(w http.ResponseWriter, r *http.Request) {
	putheader(w, r, "Comparison Query")
	putform(w, "/combos", "Enter regular expressions:",
		nCompare, nil, nSuggest, nil)
	tMultiEx.Execute(w, multixamples)
	putfooter(w, r)
}

var tMultiEx = template.Must(template.New("multixamples").Parse(
	`<P>Or choose one of these predefined sets:{{range .}}
<form action="/combos" method="post"><div>{{range $i, $s := .Exprs}}
<input type="hidden" name="x{{$i}}" value="{{$s}}">
{{end}}
<button class=link>{{.Caption}}</button></div></form>{{end}}
`))

var multixamples = []struct {
	Caption string
	Exprs   []string
}{
	{"Binary integers", []string{
		`[01]*`, `[01]+`, `0|1[01]*`, `(0|1(01*0)*1)*`, `1?(01)*0?`}},
	{"Decimal numbers", []string{
		`[1-9]\d*`, `0|-?[1-9]\d*`,
		`\d+(\.\d+)?`, `\d*\.\d+|\d+\.\d*`,
		`-?(0|[1-9]\d*)(\.\d+)?([eE][+-]?\d+)?`}},
	{"Times of the Day", []string{
		`\d?\d:\d\d`,
		`\d\d:\d\d`,
		`[012][0-9]:[0-5][0-9]`,
		`([01][0-9]|2[0-3]):[0-5][0-9]`,
		`(0[1-9]|1[012]):[0-5][0-9]`,
		`(0?[1-9]|1[012]):[0-5][0-9]`}},
	{"US Telephone Numbers", []string{
		`\(\d{3}\) \d{3}-\d{4}`,
		`\([2-9]\d\d\) [2-9]\d\d-\d\d\d\d`,
		`\([2-9][01]\d\) [2-9][2-9][2-9]-\d\d\d\d`}},
	{"US Social Security Numbers", []string{
		`\d{3}-\d{2}-\d{4}`,
		`[0-8]\d\d-\d\d-\d{4}`,
		`[0-8]\d\d-(\d[1-9]|[1-9]\d)-\d{4}`,
		`([0-6]\d{2}|7([0-6]\d|7[012]))-\d\d-\d{4}`}},
}

//  combos responds to a comparison request for multiple expressions
func combos(w http.ResponseWriter, r *http.Request) {

	// must read all input before writing anything
	// first get the expressions, trimming leading/trailing blanks
	exprlist := getexprs(r)
	nx := len(exprlist)

	// then get the test strings; these are not trimmed
	testlist := make([]string, 0)
	for _, l := range testlabels {
		arg := r.FormValue(l)
		if arg != "" {
			testlist = append(testlist, arg)
		}
	}

	// parse and echo the input
	treelist := make([]rx.Node, 0, nx)
	putheader(w, r, "Compare Expressions")
	fmt.Fprintf(w, "<P>%d expressions:\n", nx)
	for i, s := range exprlist {
		fmt.Fprintf(w, "<P class=c%d><B>e%d:</B> &nbsp; %s\n",
			i, i, hx(s))
		tree, err := rx.Parse(s)
		if !showerror(w, err) {
			treelist = append(treelist, rx.Augment(tree, i))
		}
	}

	if nx > 0 && len(treelist) == nx { // if no errors
		dfa := rx.MultiDFA(treelist) // build combined DFA
		trylist := make([]string, 0) // list of strings to try
		// tests from form submission
		for _, x := range testlist {
			trylist = append(trylist, x)
		}
		trylist = append(trylist, DRAWLINE)
		// examples from DFA
		synthx := dfa.Synthesize() // synthesize from DFA
		for _, x := range synthx { // put results on list
			trylist = append(trylist, x.Example)
		}
		trylist = append(trylist, DRAWLINE)
		// examples from parse tree
		for i := 0; i < nx; i++ {
			trylist = append(trylist, rx.Specimen(treelist[i], 1))
			trylist = append(trylist, rx.Specimen(treelist[i], 2))
			trylist = append(trylist, rx.Specimen(treelist[i], 3))
			trylist = append(trylist, rx.Specimen(treelist[i], 5))
		}
		showgrid(w, dfa, nx, trylist) // show examples
	}

	fmt.Fprintln(w, "<P>&nbsp;")
	askgraph(w, "/multaut", exprlist, "Show the automata")

	fmt.Fprint(w, "<h2>Try again?</h2>")
	putform(w, "/combos", "Enter regular expressions:",
		nCompare, exprlist, nSuggest, testlist)
	putfooter(w, r)
}

//  showgrid prints a table matching exprs with specimens (skipping duplicates).
//  A DRAWLINE line in the list draws a horizontal separator in the table
func showgrid(w http.ResponseWriter, dfa *rx.DFA, nexpr int, trylist []string) {
	seen := make(map[string]bool, 0)
	fmt.Fprintf(w, "<H2>Results</H2>\n")
	fmt.Fprintf(w, "<TABLE>\n<TR>")
	for i := 0; i < nexpr; i++ {
		fmt.Fprintf(w, "<TH class=c%d>e%d</TH>", i, i)
	}
	fmt.Fprintf(w, "<TH class=\"cg leftb\">example</TH></TR>\n")
	drawline := true
	for _, s := range trylist {
		if s == DRAWLINE {
			drawline = true
		} else if !seen[s] {
			seen[s] = true
			if drawline {
				fmt.Fprintf(w, "<TR class=topsep>")
				drawline = false
			} else {
				fmt.Fprintf(w, "<TR>")
			}
			aset := dfa.Accepts(s)
			if aset == nil {
				aset = &rx.BitSet{}
			}
			n := 0
			e := -1
			for i := 0; i < nexpr; i++ {
				if aset.Test(i) {
					n++
					e = i
					fmt.Fprintf(w,
						"<TD class=c%d>\u2713</TD>",
						i) // highlighted checkmark
				} else {
					fmt.Fprintf(w, "<TD>\u2013</TD>") // -
				}
			}
			if n == 0 {
				fmt.Fprintf(w,
					"<TD class=\"error leftb\">%s</TD></TR>\n",
					hx(s))
			} else if n == 1 {
				fmt.Fprintf(w,
					"<TD class=\"c%d leftb\">%s</TD></TR>\n",
					e, hx(s))
			} else if n < nexpr {
				fmt.Fprintf(w,
					"<TD class=\"cg leftb\">%s</TD></TR>\n",
					hx(s))
			} else {
				fmt.Fprintf(w,
					"<TD class=leftb>%s</TD></TR>\n", hx(s))
			}
		}
	}
	fmt.Fprintf(w, "</TABLE>\n")
}
