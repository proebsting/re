//  form.go -- form generation

package webapp

import (
	"html/template"
	"io"
	"net/http"
	"strings"
)

//  Form configuration values.  Code changes are needed to extend beyond 10.
var (
	nCompare   = 4  // default number of comparison fields
	nSuggest   = 2  // default number of example labels
	nMaximum   = 10 // maximum useable with the code as written
	exprlabels = []string{"x0", "x1", "x2", "x3",
		"x4", "x5", "x6", "x7", "x8", "x9"}
	testlabels = []string{"t0", "t1", "t2", "t3",
		"t4", "t5", "t6", "t7", "t8", "t9"}
)

//  getexprs retrieves and trims submitted expr values
//  (if called with none by W3C validator, make one up for better validation)
func getexprs(r *http.Request) []string {
	exprlist := make([]string, 0)
	for _, l := range exprlabels {
		arg := strings.TrimSpace(r.FormValue(l))
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
//  with at least one empty field if possible
func putform(w io.Writer, target string, label string,
	nx int, exprs []string, nt int, tests []string) {
	if nx > 1 && len(exprs) < nMaximum {
		exprs = append(exprs, "")
	}
	for len(exprs) < nx {
		exprs = append(exprs, "")
	}
	if nt > 1 && len(tests) < nMaximum {
		tests = append(tests, "")
	}
	for len(tests) < nt {
		tests = append(tests, "")
	}
	tForm.Execute(w, struct {
		Target string
		Label  string
		Exprs  []string
		Tests  []string
	}{
		target, label, exprs, tests,
	})
}

var tForm = template.Must(template.New("form").Parse(`
<form action="{{.Target}}" method=post>
<div style="margin-top: 1em"><div style="float:left;">{{.Label}} &nbsp;</div>
<div style="font-size: 67%;">
<button type=button class=link onclick="clearForm(this.form);">
(clear form)</button></div></div>
<div class=reset></div>
{{range $k, $v := .Exprs}}
<div><input tabindex=1{{$k}} type=text name=x{{$k}} size=100 maxlength=1000 value="{{$v}}"></div>{{end}}
{{if len .Tests}}<div><BR>Enter examples (optional):</div>{{end}}
{{range $k, $v := .Tests}}
<div><input tabindex=2{{$k}} type=text name=t{{$k}} size=100 maxlength=1000 value="{{$v}}"></div>{{end}}
<div><input tabindex=99 type=submit value=Submit></div></form>
`))

//  genexam generates a link (actually a form) to "examine" one predefined regex
func genexam(w io.Writer, label string, expr string) {
	formlink(w, "/details", label, []string{expr})
}

//  gencomp generates a link (actually a form) to "compare" multiple regexes
func gencomp(w io.Writer, label string, exprs []string) {
	formlink(w, "/combos", label, exprs)
}

//  formlink generates a link-like form for submitting canned examples
func formlink(w io.Writer, path string, label string, exprs []string) {
	tFormLink.Execute(w, struct {
		Path  string
		Label string
		Exprs []string
	}{
		path, label, exprs,
	})
}

var tFormLink = template.Must(template.New("formlink").Parse(`
<form action="{{.Path}}" method="post"><div>{{range $i, $s := .Exprs}}
<input type="hidden" name="x{{$i}}" value="{{$s}}">{{end}}
<button class=link>{{.Label}}</button></div></form>
`))
