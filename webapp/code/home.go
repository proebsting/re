//  home.go -- generate "home" page

package webapp

import (
	"fmt"
	"net/http"
	"rx"
)

var ex1 = `\d+(\.\d*)?(e\d+)?`
var ex3 = []string{`\d+\.\d*`, `\d+\.\d+`, `\d*\.\d+`}

//  about generates a page describing and crediting the website
func home(w http.ResponseWriter, r *http.Request) {
	putheader(w, r, "Home")
	fmt.Fprint(w, `
<P> This website provides some simple tools for experimenting with
“classic”
<A HREF="http://en.wikipedia.org/wiki/Regular_expression">regular
expressions</A> as described in computer science
textbooks and implemented in the early versions of 
<A HREF="http://en.wikipedia.org/wiki/Grep"><CITE>grep</CITE></A>.
These tools are an outgrowth of research exploring automata that
implement multiple formal languages simultaneously.
`)
	refexamine(w)
	refcompare(w)
	putfooter(w, r)
}

// refexamine advertises the Examine page
func refexamine(w http.ResponseWriter) {
	fmt.Fprint(w, `
<P> On the <A HREF="/examine">Examine</A> page you can enter
a regular expression to generate several kinds of data:
parse trees, synthetic examples, and the automata state lists.
Links from there produce diagrams of either the
<A HREF="http://en.wikipedia.org/wiki/Nondeterministic_finite_automaton">NFA</A>
or
<A HREF="http://en.wikipedia.org/wiki/Deterministic_finite_automaton">DFA</A>
for the language.
`)
	tree, _ := rx.Parse(ex1)
	augt := rx.Augment(tree, 0)
	fmt.Fprintf(w, `
<DIV CLASS=inblock>Regular Expression: %s
<BR>Augmented Parse Tree: %s
<BR class=xleading>Examples:<BR>`, hx(ex1), hx(augt))
	genexamples(w, augt, 0)
	genexamples(w, augt, 2)
	genexamples(w, augt, 4)
	fmt.Fprint(w, `
</DIV>
`)
}

// refcompare advertises the Compare page
func refcompare(w http.ResponseWriter) {
	fmt.Fprintf(w, `
<P> The <A HREF="/compare">Compare</A> page accepts multiple expressions
and shows how their languages overlap or differ.
The results page shows synthesized examples and indicates which expressions
they match.
You can also submit your own examples for testing.
Again, there are links to produce automata diagrams.
<DIV CLASS="inblock xleading">
%d expressions:
`, len(ex3))
	treelist := lpxc(w, ex3)
	dfa := rx.MultiDFA(treelist)
	synthx := dfa.Synthesize()
	trylist := make([]string, 0)
	for _, x := range synthx {
		trylist = append(trylist, x.Example)
	}
	fmt.Fprintln(w, `<BR><BR>`)
	showgrid(w, dfa, len(treelist), trylist)
	fmt.Fprintln(w, `</DIV>`)
}
