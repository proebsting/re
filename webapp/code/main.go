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

//  init registers URLs for dispatching and sets a random seed
func init() {
	http.HandleFunc("/", examine)        // anything unmatched
	http.HandleFunc("/examine", examine) // examine.go
	http.HandleFunc("/details", details) // examine.go
	http.HandleFunc("/drawDFA", drawDFA) // draw.go
	http.HandleFunc("/drawNFA", drawNFA) // draw.go
	http.HandleFunc("/compare", compare) // compare.go
	http.HandleFunc("/combos", combos)   // compare.go
	http.HandleFunc("/multaut", multaut) // graph.go
	http.HandleFunc("/syntax", syntax)   // syntax.go
	http.HandleFunc("/about", about)     // about.go
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
	if pe, ok := err.(*rx.ParseError); ok {
		fmt.Fprintf(w, "<P class=error><B>Error:</B> %s\n", pe.Message)
		fmt.Fprintf(w, "<BR>In expression: %s\n", hx(pe.BadExpr))
	} else {
		fmt.Fprintf(w, "<P class=error><B>Error:</B> %s\n", hx(err))
	}
	return true
}
