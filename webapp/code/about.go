//  about.go -- generate "about" page

package webapp

import (
	"html/template"
	"net/http"
)

//  about generates a page describing and crediting the website
func about(w http.ResponseWriter, r *http.Request) {
	putheader(w, r, "About")
	tAbout.Execute(w, nil)
	putfooter(w, r)
}

var tAbout = template.Must(template.New("about").Parse(`
<P> This website presents work by:
<P> <A HREF="http://www.cs.arizona.edu/~proebsting">Todd Proebsting</A>
<BR> <A HREF="http://www.cs.arizona.edu/~gmt">Gregg Townsend</A>
<BR> Jasmin Uribe
<P> <A HREF="http://www.cs.arizona.edu/">Department of Computer Science</A>
<BR> <A HREF="http://www.arizona.edu/">The University of Arizona</A>
<BR> <A HREF="http://www.cs.arizona.edu/camera/">Tucson, Arizona, USA</A>
<P> The website uses our own custom regular expression software written in the
<A HREF="http://golang.org/">Go</A> programming language.
<P> Graph drawing uses the <A HREF="http://www.graphviz.org">Graphviz</A>
layout package as ported to JavaScript by
<A HREF="https://github.com/mdaines/viz.js/blob/master/README.txt">Viz.js</A>
using <A HREF="https://github.com/kripken/emscripten/wiki">emscripten</A>.
This means that Graphviz runs in your own browser to lay out and draw
the graphs!
The first graph fetches a 2.5 MB script, but subsequent graphs draw quickly.
<P> The “Download” buttons for saving graphs use
cutting-edge browser features.  If they don't seem to do anything, look
in your <CODE>Downloads</CODE> folder for <CODE>NFA.svg</CODE> or
<CODE>DFA.svg</CODE> (Firefox); <CODE>download.svg</CODE> (Chrome or Opera);
or in a separate window, which can then be saved manually (Safari).
`))
