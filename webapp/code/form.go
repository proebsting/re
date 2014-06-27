//  form.go -- form generation

package webapp

import (
	"fmt"
	"io"
	"net/http"
	"rx"
	"strings"
)

//  Form configuration values.
var (
	nExpr    = 4                    // default number of comparison fields
	nTest    = 2                    // default number of example labels
	baseExpr = 100                  // base for numbering expressions
	baseTest = 200                  // base for numbering test cases
	maxExpr  = len(rx.AcceptLabels) // limit on exprs we can label
	maxTest  = maxExpr              // arbitrarily make this same
)

//  getexprs retrieves and trims submitted expr values
//  (if called with none by W3C validator, make one up for better validation)
func getexprs(r *http.Request) []string {
	exprlist := make([]string, 0)
	for i := 0; i < maxExpr; i++ {
		arg := r.FormValue(fmt.Sprintf("v%d", baseExpr+i))
		arg = strings.TrimSpace(arg)
		if arg != "" {
			exprlist = append(exprlist, arg)
		}
	}
	if len(exprlist) > 0 {
		return exprlist
	}
	if strings.HasPrefix(r.Header.Get("User-Agent"), "W3C_Validator") {
		return []string{"(a|b)*abb"}
	}
	return exprlist
}

//  putform outputs a form for submitting nx expressions and nt tests
//  with at least one empty field if possible.
//
//  changes in the HTML generated here may require corresponding changes
//  in the Javascript addslots() function.
func putform(w io.Writer, target string, label string,
	nx int, exprs []string, nt int, tests []string) {

	fmt.Fprintf(w, `
<form action="%s" method=post>
<div style="margin-top: 1em"><div style="float:left;">%s &nbsp;</div>
<div style="font-size: 67%%;">
<button type=button class=link onclick="clearForm(this.form);">
(clear form)</button>`, target, label)
	if nx > 1 && nt > 1 {
		fmt.Fprintf(w, `
<button type=button class=link onclick="addslot(%d,%d);addslot(%d,%d);">
(add slots)</button>`,
			baseExpr, maxExpr, baseTest, maxTest)
	}
	fmt.Fprintf(w, `
</div></div>
<div class=reset></div>`)
	putfields(w, exprs, nx, baseExpr, maxExpr)
	if nt > 0 {
		fmt.Fprintf(w, `
<div style="margin-top: .7ex">Enter examples (optional):</div>`)
		putfields(w, tests, nt, baseTest, maxTest)
	}
	fmt.Fprintf(w, `
<div><input tabindex=999 type=submit value=Submit></div></form>`)
}

//  putfields outputs a sequence of text input fields.
func putfields(w io.Writer, values []string, n int, base int, max int) {
	if n > 1 && len(values) < max {
		values = append(values, "")
	}
	for len(values) < n {
		values = append(values, "")
	}
	for i, v := range values {
		fmt.Fprintf(w, `
<div><input tabindex=%d type=text name=v%d size=100 maxlength=1000 value="%s"></div>`,
			base+i, base+i, hx(v))
	}
}

//  genexam generates a link (actually a form) to "examine" one predefined regex
func genexam(w io.Writer, label string, expr string) {
	formlink(w, "/details", []string{expr}, label)
}

//  gencomp generates a link (actually a form) to "compare" multiple regexes
func gencomp(w io.Writer, label string, exprs []string) {
	formlink(w, "/combos", exprs, label)
}

//  formlink generates a link-like form for submitting canned examples
func formlink(w io.Writer, path string, exprs []string, label string) {
	fmt.Fprintf(w, `
<form action="%s" method="post"><div>`, path)
	for i, v := range exprs {
		fmt.Fprintf(w, `
<input type="hidden" name="v%d" value="%s">`, baseExpr+i, hx(v))
	}
	fmt.Fprintf(w, `
<button class=link>%s</button></div></form>`, hx(label))
}
