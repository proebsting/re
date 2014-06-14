//  form.go -- form generation

package webapp

import (
	"html/template"
	"io"
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
<div><div style="float:left;">{{.Label}} &nbsp;</div>
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