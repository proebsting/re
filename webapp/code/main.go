//  main.go -- general control of the web application

package webapp

import (
	"fmt"
	"html"
	"math/rand"
	"net/http"
	"rx"
	"time"
)

//  address for sending mail (generates a mailto: link)
const MAILTO = `<script type="text/javascript">atcs('regex',null,'RegEx website');</script>`

//  init registers URLs for dispatching and sets a random seed
func init() {
	http.HandleFunc("/", home)           // home.go; anything unmatched
	http.HandleFunc("/examine", examine) // examine.go
	http.HandleFunc("/details", details) // examine.go
	http.HandleFunc("/drawDFA", drawDFA) // draw.go
	http.HandleFunc("/drawNFA", drawNFA) // draw.go
	http.HandleFunc("/compare", compare) // compare.go
	http.HandleFunc("/combos", combos)   // compare.go
	http.HandleFunc("/multaut", multaut) // graph.go
	http.HandleFunc("/syntax", syntax)   // syntax.go
	http.HandleFunc("/about", about)     // about.go
	http.HandleFunc("/contact", contact) // contact.go
	http.HandleFunc("/info", info)       // info.go
	rand.Seed(int64(time.Now().Nanosecond()))
	rx.MaxComplexity = 200 // at least twice what's needed for examples
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
	fmt.Fprintf(w, "<BR><SPAN class=error><B>Error:</B>\n")
	if pe, ok := err.(*rx.ParseError); ok {
		fmt.Fprintf(w, " %s<BR>In expression: %s",
			pe.Message, hx(pe.BadExpr))
	} else {
		fmt.Fprintf(w, "%s", hx(err))
	}
	fmt.Fprintf(w, " </SPAN>\n")
	return true
}
