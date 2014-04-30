//  This code implements the beginnings of a web application
//  using Google App Engine.
package code

import (
	"appengine"
	"bytes"
	"fmt"
	"html"
	"html/template"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"runtime"
	"rx"
	"time"
)

var examples = []struct{ Expr, Caption string }{
	{`(0|1(01*0)*1)*`, "Binary number divisible by 3"},
	{`-?(0|[1-9]\d*)(\.\d+)?([eE][+-]?\d{1,3})?`, "JSON number"},
	{`\([2-9]\d\d\) [2-9]\d\d\-\d{4}`, "US telephone number"},
	{`([0-6]\d{2}|7([0-6]\d|7[012]))-\d{2}-\d{4}`, "US social security number"},
	{`(19|20)\d\d\-(0[1-9]|1[012])\-(0[1-9]|[12]\d|3[01])`, "ISO 8601 date"},
	{`([01]\d|2[0-3]):[0-5][0-9]:[0-5][0-9]Z`, "ISO 8601 time"},
	{`\w+@\w+(\.\w+)+`, "Naive e-mail address"},
}

//  init registers URLs for dispatching and sets a random seed
func init() {
	http.HandleFunc("/info", info)
	http.HandleFunc("/response", response)
	http.HandleFunc("/", home) // anything else
	rand.Seed(int64(time.Now().Nanosecond()))
}

//  hx escapes an arbitrary stringable value for output as HTML
func hx(s interface{}) string {
	return html.EscapeString(fmt.Sprint(s))
}

//  home presents the home page
func home(w http.ResponseWriter, r *http.Request) {
	putheader(w, r, "Query")
	putform(w, "Enter a regular expression:")
	tExamples.Execute(w, examples)
	putfooter(w, r)
}

var tExamples = template.Must(template.New("header").Parse(
	`<P>Or choose one of these examples:{{range .}}
<form action="/response" method="post"><div>
<input type="hidden" name="content" value="{{.Expr}}">
<button class=link>{{.Caption}}</button></div></form>{{end}}
`))

//  response presents a response to a form submission
func response(w http.ResponseWriter, r *http.Request) {
	// must do all reading before any writing
	content := r.FormValue("content")

	putheader(w, r, "Response")
	showexpr(w, content)
	fmt.Fprint(w, "<h2>Try another?</h2>")
	putform(w, "Enter a regular expression:")
	putfooter(w, r)
}

//  showexpr generates information describing a regular expression
func showexpr(w http.ResponseWriter, s string) {
	tree, err := rx.Parse(s)
	if err != nil {
		if pe, ok := err.(*rx.ParseError); ok {
			fmt.Fprintf(w,
				"<P><B>Error:</B> %s\n<BR>In expression: %s\n",
				pe.Message, hx(pe.BadExpr))
		} else {
			fmt.Fprintf(w, "<P><B>Error:</B> %s\n", hx(err))
		}
		return
	}
	fmt.Fprintf(w, "<P>Regular Expression: %s\n", hx(s))
	// must print (or at least stringize) tree before augmenting
	fmt.Fprintf(w, "<P>Initial Parse Tree: %s\n", hx(tree))

	//#%#% currently no protection against DOS attempt from huge expr
	augt := rx.Augment(tree, 0)
	dfa := rx.BuildDFA(augt)
	dmin := dfa.Minimize()

	fmt.Fprintf(w, "<P>Augmented Tree: %s\n", hx(augt))
	fmt.Fprintf(w, "<h2>Examples</h2>\n<P>")
	showexamples(w, tree, 0)
	showexamples(w, tree, 1)
	showexamples(w, tree, 2)
	showexamples(w, tree, 3)
	showexamples(w, tree, 5)
	showexamples(w, tree, 8)

	nfaBuffer := &bytes.Buffer{}
	dfa.ShowNFA(nfaBuffer, "")
	fmt.Fprintf(w, "<h2>NFA</h2><PRE>\n%s</PRE>\n",
		hx(string(nfaBuffer.Bytes())))

	dfaBuffer := &bytes.Buffer{}
	dmin.DumpStates(dfaBuffer, "")
	fmt.Fprintf(w, "<h2>DFA</h2><PRE>\n%s</PRE>\n",
		hx(string(dfaBuffer.Bytes())))
}

const linemax = 79 // max output line length for examples

//  showexamples writes a line of specimen strings matching the expression
func showexamples(w http.ResponseWriter, tree rx.Node, maxrepl int) {
	nprinted := 0
	ncolm := 0
	for {
		s := string(tree.Example(make([]byte, 0), maxrepl))
		t := rx.Protect(s)
		ncolm += 2 + len(t)
		if nprinted > 0 && ncolm > linemax {
			break
		}
		fmt.Fprintf(w, " %s &nbsp; ", hx(t))
		nprinted++
	}
	fmt.Fprint(w, "<BR>\n")
}

//  info writes data that is useful in debugging the application
func info(w http.ResponseWriter, r *http.Request) {

	// must do all reading before any writing
	var data struct {
		Req    *http.Request
		Body   []byte
		BE     error
		Dctr   string
		GoArch string
		GoOs   string
		GoVer  string
		VID    string
		Vtime  string
	}
	data.Req = r
	data.Body, data.BE = ioutil.ReadAll(r.Body)
	data.Dctr = appengine.Datacenter()
	data.GoArch = runtime.GOARCH
	data.GoOs = runtime.GOOS
	data.GoVer = runtime.Version()
	data.VID = appengine.VersionID(appengine.NewContext(r))

	var ver int
	var bigtime int64
	fmt.Sscanf(data.VID, "%d.%d", &ver, &bigtime)
	// the following is correct for App Engine but not for SDK
	t := time.Unix(bigtime>>28, 0)
	data.Vtime = t.Format("01/02 15:04")

	putheader(w, r, "Info")
	tInfo.Execute(w, data)
	putfooter(w, r)
}

var tInfo = template.Must(template.New("header").Parse(
	`<P>Host: {{.Req.Host}}
<BR>Datacenter: {{.Dctr}} ({{.GoArch}} {{.GoOs}})
<BR>Go Version: {{.GoVer}}
<BR>App Version ID: {{.VID}} ({{.Vtime}})
<H2>Request Header</H2>
<P>{{range $k, $v := .Req.Header}}{{$k}} : {{$v}}<BR>{{end}}
<H2>Request Body</H2>
{{printf "%s" .Body}}
{{if .BE}}<P><B>BODY ERROR:</B> {{.BE}}{{end}}
`))

//  putheader outputs our standard HTML page header
func putheader(w io.Writer, r *http.Request, title string) {
	data := struct{ Prefix, Title, Favicon string }{
		"RX", title, "icon.png"}
	if r.Host == "localhost:8080" {
		data.Favicon = "itest.png"
	}
	tHeader.Execute(w, data)
}

var tHeader = template.Must(template.New("header").Parse(
	`<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 4.01//EN" "http://www.w3.org/TR/html4/strict.dtd">
<html><head>
<title>{{.Prefix}}: {{.Title}}</title>
<meta http-equiv="Content-type" content="text/html;charset=UTF-8">
<link rel="stylesheet" type="text/css" href="/static/style.css">
<link rel="icon" type="image/png" href="/static/{{.Favicon}}">
</head><body>
<h1>{{.Prefix}}: {{.Title}}</h1>
`))

//  putform outputs a form for submitting an expression
func putform(w io.Writer, label string) {
	tForm.Execute(w, label)
}

var tForm = template.Must(template.New("form").Parse(
	`<P>{{.}}
<form action="/response" method="post">
<div><input type="text" name="content" size=60 maxlength=1000></div>
<div><input type="submit" value="Submit"></div>
</form>
`))

//  putfooter outputs our standard HTML page footer
func putfooter(w io.Writer, r *http.Request) {
	tFooter.Execute(w, 1000+rand.Intn(9000))
}

var tFooter = template.Must(template.New("footer").Parse(
	`<P><BR><BR><HR>
<P style="text-align:left;">
<A title="Home" href="/">Home</a>
<span style="float:right;">
RX {{.}}
<A title="Info" href="/info">I</a>
<A title="Val" href="http://validator.w3.org/check?uri=referer&amp;ss=1">V</a>
</span> 
</body></html>
`))
