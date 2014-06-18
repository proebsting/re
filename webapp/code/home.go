//  about.go -- generate "about" page

package webapp

import (
	"html/template"
	"net/http"
)

//  about generates a page describing and crediting the website
func home(w http.ResponseWriter, r *http.Request) {
	putheader(w, r, "Home")
	tHome.Execute(w, nil)
	putfooter(w, r)
}

var tHome = template.Must(template.New("about").Parse(`
<P> This website provides some simple tools for experimenting with
“classic”
<A HREF="http://en.wikipedia.org/wiki/Regular_expression">regular
expressions</A> as described in computer science
textbooks and implemented in the early versions of 
<A HREF="http://en.wikipedia.org/wiki/Grep"><CITE>grep</CITE></A>.
These tools are an outgrowth of research exploring automata that
implement multiple formal languages simultaneously.

<P> On the <A HREF="/examine">Examine</A> page you can enter
a regular expression to generate several kinds of data:
parse trees, synthetic examples, and the automata state lists.
Links from there produce diagrams of either the
<A HREF="http://en.wikipedia.org/wiki/Nondeterministic_finite_automaton">NFA</A>
or
<A HREF="http://en.wikipedia.org/wiki/Deterministic_finite_automaton">DFA</A>
for the language.
<P> The <A HREF="/compare">Compare</A> page accepts multiple expressions
and shows how their languages overlap or differ.
The results page shows synthesized examples and indicates which expressions
they match.
You can also submit your own examples for testing.
Again, there are links to produce automata diagrams.
`))
