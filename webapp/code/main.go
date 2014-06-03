//  main.go -- general control of the web application

package webapp

import (
	"appengine"
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

//  init registers URLs for dispatching and sets a random seed
func init() {
	http.HandleFunc("/", examine)        // anything unmatched
	http.HandleFunc("/examine", examine) // examine.go
	http.HandleFunc("/details", details) // examine.go
	http.HandleFunc("/compare", compare) // compare.go
	http.HandleFunc("/combos", combos)   // compare.go
	http.HandleFunc("/syntax", syntax)
	http.HandleFunc("/about", about)
	http.HandleFunc("/info", info)
	rand.Seed(int64(time.Now().Nanosecond()))
	rx.MaxComplexity = 100 // set a relatively small limit for now
}

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

//  info generates a page of information useful to the developers
func info(w http.ResponseWriter, r *http.Request) {

	// must do all reading before any writing
	var data struct {
		Req     *http.Request
		Body    []byte
		BE      error
		Cookies []*http.Cookie
		Dctr    string
		GoArch  string
		GoOs    string
		GoVer   string
		VID     string
		Vtime   string
		MaxCx   int
	}
	data.Req = r
	data.Body, data.BE = ioutil.ReadAll(r.Body)
	data.Cookies = r.Cookies()
	data.Dctr = appengine.Datacenter()
	data.GoArch = runtime.GOARCH
	data.GoOs = runtime.GOOS
	data.GoVer = runtime.Version()
	data.VID = appengine.VersionID(appengine.NewContext(r))
	data.MaxCx = rx.MaxComplexity

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

	fmt.Fprint(w, "<P>")
	for i := 0; i < 10; i++ {
		fmt.Fprintf(w, "<span class=c%d>= c%d =</span>&nbsp;\n", i, i)
	}
	fmt.Fprintf(w, "<span class=cg>= cg =</span>&nbsp;\n") // .cg
	fmt.Fprintln(w)
	tInfo.Execute(w, data)
	putfooter(w, r)
}

var tInfo = template.Must(template.New("info").Parse(
	`<P>Host: {{.Req.Host}}
<BR>Datacenter: {{.Dctr}} ({{.GoArch}} {{.GoOs}})
<BR>Go Version: {{.GoVer}}
<BR>App Version ID: {{.VID}} ({{.Vtime}})
<BR><BR>MaxComplexity: {{.MaxCx}}
<H2>Request Header</H2>
<P>{{range $k, $v := .Req.Header}}{{$k}} : {{$v}}<BR>{{end}}
<H2>Cookies </H2><P>
{{range .Cookies}}{{.Name}} = {{.Value}}<BR>
{{else}}[none]{{end}}
<H2>Request Body</H2>
{{printf "%s" .Body}}
{{if .BE}}<P><B>BODY ERROR:</B> {{.BE}}{{end}}
`))

//  putheader outputs our standard HTML page header
//  and also initializes cookies
func putheader(w http.ResponseWriter, r *http.Request, title string) {
	initSession(w, r)
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
<meta http-equiv="Content-Script-Type" content="text/javascript">
<link rel="icon" type="image/png" href="/static/{{.Favicon}}">
<link rel="stylesheet" type="text/css" href="/static/style.css">
<script src="/static/scripts.js" type="text/javascript" defer></script>
</head><body>
<h1>{{.Prefix}}: {{.Title}}</h1>
`))

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

//  putform outputs a form for submitting n expressions
func putform(w io.Writer, n int, target string, label string, values []string) {
	fmt.Fprintf(w, "<form action=\"%s\" method=\"post\">\n", target)
	fmt.Fprintf(w, "<div><div style=\"float:left;\">%s &nbsp;</div>\n",
		label)
	fmt.Fprint(w, `<div style="font-size: 67%;">
<button type=button class=link onclick="clearForm(this.form);">
(clear form)</button></div></div>
<div class=reset></div>
`)
	for i := 0; i < n; i++ {
		fmt.Fprintf(w,
			"<div><input tabindex=%d type=\"text\" name=\"a%d\"",
			i+1, i)
		fmt.Fprintf(w, " size=100 maxlength=1000")
		if values != nil && i < len(values) && values[i] != "" {
			fmt.Fprintf(w, " value=\"%s\"", hx(values[i]))
		}
		fmt.Fprintf(w, "></div>\n")
	}
	fmt.Fprintln(w, `<div><input tabindex=99 type="submit" value="Submit"></div></form>`)
}

//  hx escapes an arbitrary stringable value for output as HTML
func hx(s interface{}) string {
	return html.EscapeString(fmt.Sprint(s))
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
