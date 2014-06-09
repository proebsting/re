//  main.go -- general control of the web application

package webapp

import (
	"fmt"
	"html"
	"io"
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
	http.HandleFunc("/compare", compare) // compare.go
	http.HandleFunc("/combos", combos)   // compare.go
	http.HandleFunc("/syntax", syntax)   // syntax.go
	http.HandleFunc("/about", about)     // about.go
	http.HandleFunc("/info", info)       // info.go
	rand.Seed(int64(time.Now().Nanosecond()))
	rx.MaxComplexity = 100 //#%#% set a relatively small limit for now
}

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
