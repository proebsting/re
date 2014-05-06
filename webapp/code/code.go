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
	"strings"
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

var multixamples = []struct {
	Caption string
	Exprs   []string
}{
	{"Decimal numbers", []string{
		`[1-9]\d*`, `0|-?[1-9]\d*`, `\d*\.\d+|\d+\.\d*`,
		`-?(0|[1-9]\d*)(\.\d+)?([eE][+-]?\d{1,3})?`}},
	{"Binary integers", []string{
		`[01]+`, `0|1[01]*`, `(0|1(01*0)*1)*`, `1?(01)*0?`}},
	{"US Social Security Numbers", []string{
		`\d\d\d-\d\d-\d\d\d\d`, `\d\d\d-(\d[1-9]|[1-9]\d)-\d\d\d\d`,
		`([0-6]\d{2}|7([0-6]\d|7[012]))-\d\d-\d\d\d\d`}},
}

//  init registers URLs for dispatching and sets a random seed
func init() {
	http.HandleFunc("/", examine)        // anything unmatched
	http.HandleFunc("/examine", examine) // anything unmatched
	http.HandleFunc("/compare", compare)
	http.HandleFunc("/details", details)
	http.HandleFunc("/combos", combos)
	http.HandleFunc("/syntax", syntax)
	http.HandleFunc("/about", about)
	http.HandleFunc("/info", info)
	rand.Seed(int64(time.Now().Nanosecond()))
}

//  hx escapes an arbitrary stringable value for output as HTML
func hx(s interface{}) string {
	return html.EscapeString(fmt.Sprint(s))
}

//  examine presents a query page for examining a single expression
func examine(w http.ResponseWriter, r *http.Request) {
	putheader(w, r, "Inspection Query")
	putform(w, 1, "/details", "Enter a regular expression:")
	tExamples.Execute(w, examples)
	putfooter(w, r)
}

var tExamples = template.Must(template.New("examples").Parse(
	`<P>Or choose one of these examples:{{range .}}
<form action="/details" method="post"><div>
<input type="hidden" name="a0" value="{{.Expr}}">
<button class=link>{{.Caption}}</button></div></form>{{end}}
`))

//  compare presents a page asking for multiple expressions
func compare(w http.ResponseWriter, r *http.Request) {
	putheader(w, r, "Comparison Query")
	putform(w, 5, "/combos", "Enter regular expressions:")
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

//  details responds to an inspection request for a single expression
func details(w http.ResponseWriter, r *http.Request) {
	expr := r.FormValue("a0") // must read before any writing
	putheader(w, r, "Inspect Expression")
	tree, err := rx.Parse(expr)
	if err == nil {
		fmt.Fprintf(w, "<P>Regular Expression: %s\n", hx(expr))
		// must print (or at least stringize) tree before augmenting
		fmt.Fprintf(w, "<P>Initial Parse Tree: %s\n", hx(tree))

		//#%#% currently no protection against DOS by huge expr
		augt := rx.Augment(tree, 0)
		dfa := rx.BuildDFA(augt)
		dmin := dfa.Minimize()

		fmt.Fprintf(w, "<P>Augmented Tree: %s\n", hx(augt))
		fmt.Fprintf(w, "<h2>Examples</h2>\n<P>")
		genexamples(w, tree, 0)
		genexamples(w, tree, 1)
		genexamples(w, tree, 2)
		genexamples(w, tree, 3)
		genexamples(w, tree, 5)
		genexamples(w, tree, 8)

		nfaBuffer := &bytes.Buffer{}
		dfa.ShowNFA(nfaBuffer, "")
		fmt.Fprintf(w, "<h2>NFA</h2><PRE>\n%s</PRE>\n",
			hx(string(nfaBuffer.Bytes())))

		dfaBuffer := &bytes.Buffer{}
		dmin.DumpStates(dfaBuffer, "")
		fmt.Fprintf(w, "<h2>DFA</h2><PRE>\n%s</PRE>\n",
			hx(string(dfaBuffer.Bytes())))
	} else {
		showerror(w, err)
	}

	fmt.Fprint(w, "<h2>Try another?</h2>")
	putform(w, 1, "/details", "Enter a regular expression:")
	putfooter(w, r)
}

//  showerror displays an error and returns true if its argument is not nil
func showerror(w http.ResponseWriter, err error) bool {
	if err == nil {
		return false
	}
	if pe, ok := err.(*rx.ParseError); ok {
		fmt.Fprintf(w, "<P><B>Error:</B> %s\n<BR>In expression: %s\n",
			pe.Message, hx(pe.BadExpr))
	} else {
		fmt.Fprintf(w, "<P><B>Error:</B> %s\n", hx(err))
	}
	return true
}

const linemax = 79 // max output line length for examples

//  genexamples writes a line of specimen strings matching the expression
func genexamples(w http.ResponseWriter, tree rx.Node, maxrepl int) {
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
		fmt.Fprintf(w, "<P><B>e%d:</B> &nbsp; %s\n", i, hx(s))
		tree, err := rx.Parse(s)
		if !showerror(w, err) {
			tlist = append(tlist, rx.Augment(tree, uint(i)))
		}
	}

	if n > 0 && len(tlist) == n { // if no errors
		dfa := rx.MultiDFA(tlist)  // build combined DFA
		xlist := make([]string, 0) // example list
		synthx := dfa.Synthesize() // synthesize from DFA
		for _, x := range synthx { // put results on list
			xlist = append(xlist, x.Example)
		}
		showgrid(w, dfa, n, xlist) // show examples
	}
	fmt.Fprint(w, "<h2>Try again?</h2>")
	putform(w, 5, "/combos", "Enter regular expressions:")
	putfooter(w, r)
}

//  showgrid prints a table showing which expr matched which example
func showgrid(w http.ResponseWriter, dfa *rx.DFA, nexpr int, xlist []string) {
	fmt.Fprintf(w, "<H2>Results</H2>\n")
	fmt.Fprintf(w, "<TABLE>\n<TR>")
	for i := 0; i < nexpr; i++ {
		fmt.Fprintf(w, "<TH>e%d</TH>", i)
	}
	fmt.Fprintf(w, "<TH>example</TH></TR>\n")
	for _, s := range xlist {
		fmt.Fprintf(w, "<TR>")
		aset := dfa.Accepts(s)
		for i := 0; i < nexpr; i++ {
			if aset.Test(uint(i)) {
				fmt.Fprintf(w, "<TD>\u2713</TD>") // ckmark
			} else {
				fmt.Fprintf(w, "<TD>\u2013</TD>") // ndash
			}
		}
		fmt.Fprintf(w, "<TD>%s</TD></TR>\n", hx(s))
	}
	fmt.Fprintf(w, "</TABLE>\n")
}

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
`))

//  syntax generates a page outlining the accepted syntax
func syntax(w http.ResponseWriter, r *http.Request) {
	putheader(w, r, "Syntax")
	tSyntax.Execute(w, r)
	putfooter(w, r)
}

var tSyntax = template.Must(template.New("syntax").Parse(`
<P>We implement traditional regular expressions with a few simple extensions.
<P>The following forms are handled:<PRE>
      abc  a|b|c  a(b|c)d
      a?  b*  c+  d{m,n}
      \a \e \f \n \r \t \v \046 \xF7 \u03A8
      .  \d \s \w \D \S \W
      [abc]  [^abc]  [a-c]  [\x]
</PRE>
<P>We ignore the Perl non-capturing submatch form <CODE>?:</CODE>,
but other <CODE>(?</CODE> forms are errors.
<P>All expressions are &ldquo;anchored&rdquo;.
An initial <CODE>^</CODE> and/or final <CODE>$</CODE> is ignored.
`))

//  info generates a page of information useful to the developers
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
	if strings.HasPrefix(r.Host, "localhost:") {
		// don't know how to decode this in SDK environment
		data.Vtime = "?!"
	} else {
		// decode VID to match with appengine admin logs
		t := time.Unix(bigtime>>28, 0)
		data.Vtime = t.Format("01/02 15:04")
	}

	putheader(w, r, "Info")
	tInfo.Execute(w, data)
	putfooter(w, r)
}

var tInfo = template.Must(template.New("info").Parse(
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

//  putform outputs a form for submitting n expressions
func putform(w io.Writer, n int, target string, label string) {
	fmt.Fprintf(w, "<P>%s\n<form action=\"%s\" method=\"post\">",
		label, target)
	for i := 0; i < n; i++ {
		fmt.Fprintf(w, `
<div><input type="text" name="a%d" size=60 maxlength=1000></div>`, i)
	}
	fmt.Fprintln(w, `
<div><input type="submit" value="Submit"></div>
</form>`)
}

//  putfooter outputs our standard HTML page footer
func putfooter(w io.Writer, r *http.Request) {
	tFooter.Execute(w, 1000+rand.Intn(9000))
}

var tFooter = template.Must(template.New("footer").Parse(
	`<P><BR><BR><HR>
<P style="text-align:left;">
<A title="Home" href="/">Home</a>
| <A title="Examine" href="/examine">Examine</a>
| <A title="Compare" href="/compare">Compare</a>
| <A title="Syntax" href="/syntax">Syntax</a>
| <A title="About" href="/about">About</a>
<span style="float:right;">
RX {{.}}
<A title="Info" href="/info">I</a>
<A title="Val" href="http://validator.w3.org/check?uri=referer&amp;ss=1">V</a>
</span> 
</body></html>
`))
