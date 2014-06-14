//  draw.go -- code for drawing a DFA or NFA

package webapp

import (
	"fmt"
	"html/template"
	"net/http"
	"rx"
	"strings"
)

//  askgraph generates a "draw the graph" link
func askgraph(w http.ResponseWriter, path string, exprlist []string, label string) {
	tAskGraph.Execute(w,
		struct {
			Path     string
			ExprList []string
			Label    string
		}{
			path, exprlist, label})
}

var tAskGraph = template.Must(template.New("askgraph").Parse(
	`<form action="{{.Path}}" method="post"><div>
{{range $k, $v := .ExprList}}{{if $v}}
<input type="hidden" name="x{{$k}}" value="{{$v}}">{{end}}{{end}}
<button class=link>{{.Label}}</button></div></form>
`))

//  drawDFA draws a DFA.
func drawDFA(w http.ResponseWriter, r *http.Request) {
	draw(w, r, "DFA")
}

//  drawNFA draws a DFA.
func drawNFA(w http.ResponseWriter, r *http.Request) {
	draw(w, r, "NFA")
}

//  draw produces a Dot file for rendering a DFA or NFA in the user's browser.
func draw(w http.ResponseWriter, r *http.Request, which string) {
	// must read all input before writing anything
	exprlist := make([]string, 0)
	treelist := make([]rx.Node, 0)
	for _, l := range exprlabels {
		arg := strings.TrimSpace(r.FormValue(l))
		if arg != "" {
			exprlist = append(exprlist, arg)
			tree, err := rx.Parse(arg)
			if err == nil {
				treelist = append(treelist,
					rx.Augment(tree, len(treelist)))
			}
		}
	}

	putheader(w, r, which+" Graph")
	if len(treelist) == 0 {
		fmt.Fprintln(w, `<P>[no valid expressions found]`)
	} else {
		dfa := rx.MultiDFA(treelist)
		dmin := dfa.Minimize()

		fmt.Fprintf(w, "<P>\n")
		for _, e := range exprlist {
			fmt.Fprintf(w, "%s<BR>\n", hx(e))
		}
		fmt.Fprintln(w, `<script type="text/vnd.graphviz" id="graph">`)
		if which == "NFA" {
			dmin.GraphNFA(w, "")
		} else {
			dmin.ToDot(w, "")
		}
		fmt.Fprintln(w, `</script>`)

		tDraw.Execute(w, which)
	}

	putfooter(w, r)
}

var tDraw = template.Must(template.New("draw").Parse(`
<script type="text/javascript" src="/static/viz.js"></script>
<script type="text/javascript">
function download(filename, mimetype, data) {
  var pom = document.createElement('a');
  pom.setAttribute('href',
    'data:' + mimetype + ';charset=utf-8,' + encodeURIComponent(data));
  pom.setAttribute('download', filename);
  document.body.appendChild(pom)
  pom.click();
  document.body.removeChild(pom)
}
var dot = document.getElementById("graph").innerHTML;
var svg = Viz(dot, "svg")
document.body.innerHTML += svg;
</script>
<P>
<input type="button" value="Download SVG image"
onclick="download('{{.}}.svg', 'image/svg+xml', svg);" />
<input type="button" value="Download Graphviz commands"
onclick="download('{{.}}.gv', 'text/vnd.graphviz', dot);" />
`))
