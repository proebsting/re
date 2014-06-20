//  draw.go -- code for drawing a DFA or NFA

package webapp

import (
	"fmt"
	"html/template"
	"net/http"
	"rx"
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

	exprlist := getexprs(r) // must load data before writing anything
	nx := len(exprlist)

	putheader(w, r, which+" Graph") // write page header
	fmt.Fprintln(w, "<P class=xleading>")

	treelist := make([]rx.Node, 0)
	for i, e := range exprlist {
		if nx > 1 {
			fmt.Fprintf(w, "%c. &nbsp; ", rx.AcceptLabels[i])
		}
		fmt.Fprintf(w, "%s<BR>\n", hx(e))
		tree, err := rx.Parse(e)
		if !showerror(w, err) {
			treelist = append(treelist, rx.Augment(tree, i))
		}
	}

	if nx > 0 && len(treelist) == nx { // if no errors
		dfa := rx.MultiDFA(treelist) // build combined DFA
		dmin := dfa.Minimize()       // minimize it

		fmt.Fprintln(w, `<script type="text/vnd.graphviz" id="graph">`)
		if which == "NFA" {
			dmin.GraphNFA(w, "")
		} else if nx == 1 {
			dmin.ToDot(w, "", "")
		} else {
			which = "Multi"
			dmin.ToDot(w, "", rx.AcceptLabels)
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
<P>{{if eq . "NFA"}} The unlabeled node is the start state.
{{else}} State s0 is the start state.{{end}}
{{if eq . "Multi"}} Double octagons are acceptance states, with capital
letters indicating <EM>which</EM> expressions are accepted.
{{else}} Double circles are acceptance states.{{end}}
Edge labels indicate input characters or classes of characters.
<P>
<input type="button" value="Download SVG image"
onclick="download('{{.}}.svg', 'image/svg+xml', svg);">
<input type="button" value="Download Graphviz commands"
onclick="download('{{.}}.gv', 'text/vnd.graphviz', dot);">
`))
