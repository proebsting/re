package code

import (
	"appengine"
	"fmt"
	"html"
	//"html/template"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"rx"
	"strconv"
	"time"
)

func init() {
	http.HandleFunc("/info", info)
	http.HandleFunc("/response", response)
	http.HandleFunc("/n/", n31)	// 3n+1
	http.HandleFunc("/", home)		// anything else 
	rand.Seed(int64(time.Now().Nanosecond()))
}

func home(w http.ResponseWriter, r *http.Request) {
	header(w, r, "Query")
	fmt.Fprint(w, `
<P><form action="/response" method="post">
<div><input type="text" name="content" size=30 maxlength=100></div>
<div><input type="submit" value="Submit"></div>
</form>
`)
	footer(w, r)
}

func response(w http.ResponseWriter, r *http.Request) {
	// must do all reading before any writing
	content := html.EscapeString(r.FormValue("content"))

        header(w, r, "Response")
	fmt.Fprintf(w, "<P>Query was: %s\n", content)
	footer(w, r)
}

func n31(w http.ResponseWriter, r *http.Request) {
	header(w, r, "3N + 1")
	bs := &rx.BitSet{}
	n, _ := strconv.Atoi(r.URL.Path[3:])
	if n == 0 {
		n = rand.Intn(100)
	}
	fmt.Fprintf(w, "<P><B>%d</B>", n)
	for n > 1 {
		bs.Set(uint(n))
		if n&1 == 1 {
			n = 3*n + 1
		} else {
			n = n / 2
		}
		fmt.Fprintf(w, " %d", n)
	}
	fmt.Fprintf(w, "\n<BR><BR>%s\n", bs)
	footer(w, r)
}

func info(w http.ResponseWriter, r *http.Request) {
	// must do all reading before any writing
	var body []byte
	var berr error
	if r.Body != nil {
		body, berr = ioutil.ReadAll(r.Body)
	}

	header(w, r, "Info")
	fmt.Fprintf(w, "<P>Host: %#v\n", r.Host)
	fmt.Fprint(w, "<P>")
	for k, v := range r.Header {
		fmt.Fprintf(w, "%s : %s<BR>\n", k, v)
	}

	if berr != nil {
		fmt.Fprintf(w, "<P>BODY ERROR: %s\n", berr)
	}
	fmt.Fprintf(w, "<P>Body:<BR>%s\n", html.EscapeString(string(body)))

	fmt.Fprintf(w, "<P>Datacenter: %s\n", appengine.Datacenter())

	vid := appengine.VersionID(appengine.NewContext(r))
	var ver int
	var bigtime int64
	fmt.Sscanf(vid, "%d.%d", &ver, &bigtime)
	// the following is correct for App Engine but not for SDK
	t := time.Unix(bigtime >> 28, 0)
	fmt.Fprintf(w, "<P>App Version ID: %s (%s)\n", vid,
		t.Format("01/02 15:04"))
	 

	footer(w, r)
}

func header(w io.Writer, r *http.Request, title string) {
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

func footer(w io.Writer, r *http.Request) {
	fmt.Fprintf(w, `
<P><HR><P>
<P> RX [%d] :
<a title="Home" href="/home">Home</a> |
<a title="Info" href="/info">Info</a> |
<a title="Validate" href="http://validator.w3.org/check?uri=referer&amp;ss=1"> 
Validate</a> |
<a title="27" href="/n/27">27</a>
</body></html>
`, 1000+rand.Intn(9000))
}
