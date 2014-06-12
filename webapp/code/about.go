//  about.go -- generate "about" page

package webapp

import (
	"html/template"
	"net/http"
)

//  about generates a page describing and crediting the website
func about(w http.ResponseWriter, r *http.Request) {
	putheader(w, r, "About")
	tAbout.Execute(w, r)
	putfooter(w, r)
}

var tAbout = template.Must(template.New("about").Parse(`
<P> This website is a work in progress by:
<P> <A HREF="http://www.cs.arizona.edu/~proebsting">Todd Proebsting</A>
<BR> <A HREF="http://www.cs.arizona.edu/~gmt">Gregg Townsend</A>
<BR> Jasmin Uribe
<P> <A HREF="http://www.cs.arizona.edu/">Department of Computer Science</A>
<BR> <A HREF="http://www.arizona.edu/">The University of Arizona</A>
<BR> Tucson, Arizona, USA
<P>
Graph drawing uses the <A HREF="http://www.graphviz.org">Graphviz</A>
layout package as ported to JavaScript by
<A HREF="https://github.com/mdaines/viz.js/blob/master/README.txt">Viz.js</A>
using <A HREF="https://github.com/kripken/emscripten/wiki">emscripten</A>.
This means that Graphviz is running in your own browser to draw the graphs!
`))
