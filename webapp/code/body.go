//  body.go -- generate headers and footers for all pages

package webapp

import (
	"html/template"
	"io"
	"math/rand"
	"net/http"
)

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