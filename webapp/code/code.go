//  This code implements the beginnings of a web application
//  using Google App Engine.
package code

import (
	"appengine"
	"fmt"
	"html"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"rx"
	"time"
)

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
	putfooter(w, r)
}

//  response presents a response to a form submission
func response(w http.ResponseWriter, r *http.Request) {
	// must do all reading before any writing
	content := r.FormValue("content")

	putheader(w, r, "Response")
	showexpr(w, content)
	fmt.Fprintf(w, "<h2>Try another?</h2>")
	putform(w, "Enter a regular expression:")
	putfooter(w, r)
}

//  showexpr generates information describing a regular expression
func showexpr(w http.ResponseWriter, s string) {
	tree, err := rx.Parse(s)
	if err != nil {
		fmt.Fprintf(w, "<P>ERROR: %s\n", hx(err))
		return
	}
	fmt.Fprintf(w, "<P>Expression: %s\n", s)
	fmt.Fprintf(w, "<P>Parse Tree: %s\n", hx(tree))

	//#%#% currently no protection against DOS attempt from huge expr
	//#%#% likewise input strings are not fully sanitized
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
	fmt.Fprintf(w, "<h2>NFA</h2>\n<PRE>\n")
	dfa.ShowNFA(w, "")
	fmt.Fprintf(w, "</PRE>\n")
	fmt.Fprintf(w, "<h2>DFA</h2>\n<PRE>\n")
	dmin.DumpStates(w, "")
	fmt.Fprintf(w, "</PRE>\n")
}

const linemax = 79 // max output line length for examples

//  showexamples writes a line of specimen strings matching the expression
func showexamples(w http.ResponseWriter, tree rx.Node, maxrepl int) {
	fmt.Fprintf(w, "<BR>")
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
	fmt.Fprintf(w, "\n")
}

//  info writes data that is useful in debugging the application
func info(w http.ResponseWriter, r *http.Request) {
	// must do all reading before any writing
	var body []byte
	var berr error
	if r.Body != nil {
		body, berr = ioutil.ReadAll(r.Body)
	}

	putheader(w, r, "Info")
	fmt.Fprintf(w, "<P>Host: %#v\n", r.Host)
	fmt.Fprint(w, "<P>")
	for k, v := range r.Header {
		fmt.Fprintf(w, "%s : %s<BR>\n", k, v)
	}

	if berr != nil {
		fmt.Fprintf(w, "<P>BODY ERROR: %s\n", berr)
	}
	fmt.Fprintf(w, "<P>Body:<BR>%s\n", hx(body))

	fmt.Fprintf(w, "<P>Datacenter: %s\n", appengine.Datacenter())

	vid := appengine.VersionID(appengine.NewContext(r))
	var ver int
	var bigtime int64
	fmt.Sscanf(vid, "%d.%d", &ver, &bigtime)
	// the following is correct for App Engine but not for SDK
	t := time.Unix(bigtime>>28, 0)
	fmt.Fprintf(w, "<P>App Version ID: %s (%s)\n", vid,
		t.Format("01/02 15:04"))

	putfooter(w, r)
}

//  putheader outputs our standard HTML page header
func putheader(w io.Writer, r *http.Request, title string) {
	prefix := "RX"
	favicon := "icon.png"
	if r.Host == "localhost:8080" {
		favicon = "ilocal.png"
	}
	fmt.Fprintf(w, `
<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 4.01//EN" "http://www.w3.org/TR/html4/strict.dtd">
<html><head>
<title>%s: %s</title>
<meta http-equiv="Content-type" content="text/html;charset=UTF-8">
<link rel="stylesheet" type="text/css" href="/static/style.css">
<link rel="icon" type="image/png" href="/static/%s">
</head><body>
<h1>%s: %s</h1>
`, prefix, title, favicon, prefix, title)
}

//  putform outputs a form for submitting an expression
func putform(w io.Writer, label string) {
	fmt.Fprintf(w, `
<P>%s
<P><form action="/response" method="post">
<div><input type="text" name="content" size=60 maxlength=1000></div>
<div><input type="submit" value="Submit"></div>
</form>
`, label)
}

//  putfooter outputs our standard HTML page footer
func putfooter(w io.Writer, r *http.Request) {
	fmt.Fprintf(w, `
<P><BR><BR><HR>
<P style="text-align:left;">
<A title="Home" href="/">Home</a>
<span style="float:right;">
RX %d
<A title="Info" href="/info">I</a>
<A title="Val" href="http://validator.w3.org/check?uri=referer&amp;ss=1">V</a>
</span> 
</body></html>
`, 1000+rand.Intn(9000))
}
