// compare.go -- code for comparing multiple regular expressions

package webapp

import (
	"fmt"
	"net/http"
	"rx"
)

var DRAWLINE = "\x7F\x7F" // special flag for separator in grid
var NCOLORS = 9           // number of defined cn color classes in style.css

//  compare presents a page asking for multiple expressions
func compare(w http.ResponseWriter, r *http.Request) {
	putheader(w, r, "Comparison Query")
	fmt.Fprintln(w, `<P> Here you can specify multiple expressions
to see how their languages overlap or differ.
The results page shows synthesized examples and indicates which expressions
they match.  You can also submit your own examples for testing.`)
	putform(w, "/combos", "Enter regular expressions:",
		nExpr, nil, nTest, nil)
	fmt.Fprintln(w, `<P>Or choose one of these predefined sets:`)
	for _, x := range multixamples {
		gencomp(w, x.Caption, x.Exprs)
	}
	putfooter(w, r)
}

var multixamples = []struct {
	Caption string
	Exprs   []string
}{
	{HomeCmpLabel, HomeCompare},
	{"Decimal numbers", []string{
		`[1-9]\d*`, `0|-?[1-9]\d*`,
		`\d+(\.\d+)?`, `\d*\.\d+|\d+\.\d*`,
		`-?(0|[1-9]\d*)(\.\d+)?([eE][+-]?\d+)?`}},
	{"Binary integers", []string{
		`[01]*`, `[01]+`, `0|1[01]*`, `(0|1(01*0)*1)*`, `1?(01)*0?`}},
	{"Times of the Day", []string{
		`\d?\d:\d\d`,
		`\d\d:\d\d`,
		`[012][0-9]:[0-5][0-9]`,
		`([01][0-9]|2[0-3]):[0-5][0-9]`,
		`(0[1-9]|1[012]):[0-5][0-9]`,
		`(0?[1-9]|1[012]):[0-5][0-9]`}},
	{"US Telephone Numbers", []string{
		`\([2-9][0-1][1-9]\) [2-9][2-9][1-9]-\d\d\d\d`, // 1950s
		`\([2-9][0-8][0-9]\) [2-9]\d\d-\d\d\d\d`,       // 1990s
		`\([2-9][0-9][0-9]\) [2-9]\d\d-\d\d\d\d`,       // future
		`\(\d{3}\) \d{3}-\d{4}`}},                      // unrestricted
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
	for i := 0; i < maxTest; i++ {
		arg := r.FormValue(fmt.Sprintf("v%d", baseTest+i))
		if arg != "" {
			testlist = append(testlist, arg)
		}
	}

	// parse and echo the input
	putheader(w, r, "Compare Expressions")
	fmt.Fprintf(w, "<P class=xleading>%d expressions:\n", nx)
	treelist := lpxc(w, exprlist)

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
		// DON'T separate parse tree examples
		// (confusing in the the absence of a published explanation)
		// trylist = append(trylist, DRAWLINE)
		// examples from parse tree
		for i := 0; i < nx; i++ {
			trylist = append(trylist, rx.Specimen(treelist[i], 1))
			trylist = append(trylist, rx.Specimen(treelist[i], 2))
			trylist = append(trylist, rx.Specimen(treelist[i], 3))
			trylist = append(trylist, rx.Specimen(treelist[i], 5))
		}
		fmt.Fprintf(w, "<H2>Results</H2>\n")
		showgrid(w, dfa, nx, trylist) // show examples
	}

	fmt.Fprintln(w, "<P>&nbsp;")
	formlink(w, "/multaut", exprlist, "Show automata states")
	formlink(w, "/drawNFA", exprlist, "Draw the NFA")
	formlink(w, "/drawDFA", exprlist, "Draw the DFA")

	fmt.Fprint(w, "<h2>Try again?</h2>")
	putform(w, "/combos", "Enter regular expressions:",
		nExpr, exprlist, nTest, testlist)
	putfooter(w, r)
}

//  lpxc -- list and parse expressions for comparison, returning augtree list
func lpxc(w http.ResponseWriter, exprlist []string) []rx.Node {
	treelist := make([]rx.Node, 0)

	for i, s := range exprlist {
		fmt.Fprintf(w,
			"<SPAN class=%s><BR><B>%c:</B> &nbsp; %s</SPAN>\n",
			colorname(i), rx.AcceptLabels[i], hx(s))
		tree, err := rx.Parse(s)
		if !showerror(w, err) {
			treelist = append(treelist, rx.Augment(tree, i))
		}
	}
	return treelist
}

//  showgrid prints a table matching exprs with specimens (skipping duplicates).
//  A DRAWLINE line in the list draws a horizontal separator in the table
func showgrid(w http.ResponseWriter, dfa *rx.DFA, nexpr int, trylist []string) {
	seen := make(map[string]bool, 0)
	fmt.Fprintf(w, "<TABLE>\n<TR>")
	for i := 0; i < nexpr; i++ {
		fmt.Fprintf(w, "<TH class=%s>%c</TH>",
			colorname(i), rx.AcceptLabels[i])
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
					fmt.Fprintf(w, // highlighted checkmark
						"<TD class=%s>\u2713</TD>",
						colorname(i))
				} else {
					fmt.Fprintf(w, "<TD>\u2013</TD>") // -
				}
			}
			class := "cg"
			switch n {
			case 0:
				class = "error"
			case 1:
				class = fmt.Sprintf("%s", colorname(e))
			case nexpr:
				class = "cw"
			}
			fmt.Fprintf(w,
				"<TD class=\"leftb %s\">%s</TD></TR>\n",
				class, hx(s))
		}
	}
	fmt.Fprintf(w, "</TABLE>\n")
}

//  colorname returns the CSS class name associated with a color
func colorname(i int) string {
	return fmt.Sprintf("c%d", i%NCOLORS)
}
