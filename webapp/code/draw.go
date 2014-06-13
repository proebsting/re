//  draw.go -- code for drawing a DFA or NFA

package webapp

import (
	"fmt"
	"html/template"
	"net/http"
	"rx"
	"strings"
)

//  drawDFA draw a DFA.
func drawDFA(w http.ResponseWriter, r *http.Request) {
	draw(w, r, "DFA")
}

//  drawNFA draw a DFA.
func drawNFA(w http.ResponseWriter, r *http.Request) {
	draw(w, r, "NFA")
}

//  draw produces a Dot file for rendering a DFA or NFA in the user's browser.
func draw(w http.ResponseWriter, r *http.Request, which string) {
	expr := r.FormValue("x0")      // must read before any writing
	expr = strings.TrimSpace(expr) // trim leading/trailing blanks

	putheader(w, r, which+" Graph")
	tree, err := rx.Parse(expr)
	if err != nil {
		showerror(w, err)
		putfooter(w, r)
		return
	}
	augt := rx.Augment(tree, 0)
	dfa := rx.BuildDFA(augt)
	dmin := dfa.Minimize()

	fmt.Fprintf(w, "<P>%s\n<P>\n", hx(expr))
	fmt.Fprintln(w, `<script type="text/vnd.graphviz" id="graph">`)
	if which == "NFA" {
		dmin.GraphNFA(w, "")
	} else {
		dmin.ToDot(w, "")
	}
	fmt.Fprintln(w, `</script>`)

	tDraw.Execute(w, which)

	putfooter(w, r)
}

var tDraw = template.Must(template.New("draw").Parse(`
<script type="text/javascript" src="/static/viz.js"></script>
<script type="text/javascript">
function pretect(s) {
	return "<pre>" + s.replace(/\&/g, "&amp;").replace(/\"/g, "&quot;").
		replace(/</g, "&lt;").replace(/>/g, "&gt;") + "</pre>";
}
function render(id) {
	try {
		dot = document.getElementById(id).innerHTML;
		return Viz(dot, "svg");
	} catch(e) {
		return pretect(e.toString());
	}
}
function download(filename, data) {
	var pom = document.createElement('a');
	pom.setAttribute('href',
		'data:image/xvg+xml;charset=utf-8,' + encodeURIComponent(data));
	pom.setAttribute('download', filename);
	document.body.appendChild(pom)
	pom.click();
	document.body.removeChild(pom)
}

var svg = render("graph")
document.body.innerHTML += svg;
</script>
<P>
<input type="button" value="Download SVG file"
onclick="download('{{.}}.svg', svg);" />
`))
