// compare.go -- code for comparing multiple regular expressions

package webapp

import (
	"fmt"
	"html/template"
	"net/http"
	"rx"
)

//  compare presents a page asking for multiple expressions
func compare(w http.ResponseWriter, r *http.Request) {
	putheader(w, r, "Comparison Query")
	putform(w, 5, "/combos", "Enter regular expressions:", nil)
	tMultiEx.Execute(w, multixamples)
	putfooter(w, r)
}

var tMultiEx = template.Must(template.New("multixamples").Parse(
	`<P>Or choose one of these predefined sets:{{range .}}
<form action="/combos" method="post"><div>{{range $i, $s := .Exprs}}
<input type="hidden" name="a{{$i}}" value="{{$s}}">
{{end}}
<button class=link>{{.Caption}}</button></div></form>{{end}}
`))

var multixamples = []struct {
	Caption string
	Exprs   []string
}{
	{"Binary integers", []string{
		`[01]+`, `0|1[01]*`, `(0|1(01*0)*1)*`, `1?(01)*0?`}},
	{"Decimal numbers", []string{
		`[1-9]\d*`, `0|-?[1-9]\d*`, `\d*\.\d+|\d+\.\d*`,
		`-?(0|[1-9]\d*)(\.\d+)?([eE][+-]?\d{1,3})?`}},
	{"Times of the Day", []string{
		`\d\d:\d\d`, `[012][0-9]:[0-5][0-9]`,
		`([01][0-9]|2[0-3]):[0-5][0-9]`,
		`(0[1-9]|1[012]):[0-5][0-9]`}},
	{"US Telephone Numbers", []string{
		`\(\d{3}\)-\d{3}-\d{4}`,
		`\([2-9]\d\d\)-[2-9]\d\d-\d\d\d\d`,
		`\([2-9][01]\d\)-[2-9][2-9][2-9]-\d\d\d\d`}},
	{"US Social Security Numbers", []string{
		`\d\d\d-\d\d-\d\d\d\d`, `\d\d\d-(\d[1-9]|[1-9]\d)-\d\d\d\d`,
		`([0-6]\d{2}|7([0-6]\d|7[012]))-\d\d-\d\d\d\d`}},
}

//  combos responds to a comparison request for multiple expressions
func combos(w http.ResponseWriter, r *http.Request) {

	// must read all input before writing anything
	labels := []string{"a0", "a1", "a2", "a3", "a4"}
	elist := make([]string, 0, 5)
	for _, l := range labels {
		arg := r.FormValue(l)
		if arg != "" {
			elist = append(elist, arg)
		}
	}
	n := len(elist)

	tlist := make([]rx.Node, 0, n)
	putheader(w, r, "Compare Expressions")
	fmt.Fprintf(w, "<P>%d expressions:\n", n)
	for i, s := range elist {
		fmt.Fprintf(w, "<P class=c%d><B>e%d:</B> &nbsp; %s\n",
			i, i, hx(s))
		tree, err := rx.Parse(s)
		if !showerror(w, err) {
			tlist = append(tlist, rx.Augment(tree, i))
		}
	}

	if n > 0 && len(tlist) == n { // if no errors
		dfa := rx.MultiDFA(tlist)  // build combined DFA
		xlist := make([]string, 0) // example list
		synthx := dfa.Synthesize() // synthesize from DFA
		for _, x := range synthx { // put results on list
			xlist = append(xlist, x.Example)
		}
		for i := 0; i < n; i++ {
			xlist = append(xlist, rx.Specimen(tlist[i], 1))
			xlist = append(xlist, rx.Specimen(tlist[i], 2))
			xlist = append(xlist, rx.Specimen(tlist[i], 3))
			xlist = append(xlist, rx.Specimen(tlist[i], 5))
		}
		showgrid(w, dfa, n, xlist) // show examples
	}
	fmt.Fprint(w, "<h2>Try again?</h2>")
	putform(w, 5, "/combos", "Enter regular expressions:", elist)
	putfooter(w, r)
}

//  showgrid prints a table matching exprs with specimens (skipping duplicates)
func showgrid(w http.ResponseWriter, dfa *rx.DFA, nexpr int, xlist []string) {
	seen := make(map[string]bool, 0)
	fmt.Fprintf(w, "<H2>Results</H2>\n")
	fmt.Fprintf(w, "<TABLE>\n<TR>")
	for i := 0; i < nexpr; i++ {
		fmt.Fprintf(w, "<TH class=c%d>e%d</TH>", i, i)
	}
	fmt.Fprintf(w, "<TH class=leftw>example</TH></TR>\n")
	for _, s := range xlist {
		if !seen[s] {
			seen[s] = true
			fmt.Fprintf(w, "<TR>")
			aset := dfa.Accepts(s)
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
			if n == 1 {
				fmt.Fprintf(w,
					"<TD class=\"c%d leftw\">%s</TD></TR>\n",
					e, hx(s))
			} else {
				fmt.Fprintf(w,
					"<TD class=leftw>%s</TD></TR>\n", hx(s))
			}
		}
	}
	fmt.Fprintf(w, "</TABLE>\n")
}
