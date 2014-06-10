//  form.go -- form generation

package webapp

import (
	"html/template"
	"io"
)

//  Form configuration values.  Code changes are needed to extend beyond 10.
var (
	exprlabels = []string{"a0", "a1", "a2", "a3", "a4", "a5"}
	testlabels = []string{"t0", "t1", "t2", "t3"}
	Ncompare   = len(exprlabels)
	Nsuggest   = len(testlabels)
)

//  putform outputs a form for submitting nx expressions and nt tests
func putform(w io.Writer, target string, label string,
	nx int, exprs []string, nt int, tests []string) {
	for len(exprs) < nx {
		exprs = append(exprs, "")
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
<div><input tabindex=1{{$k}} type=text name=a{{$k}} size=100 maxlength=1000 value="{{$v}}"></div>{{end}}
{{if len .Tests}}<div><BR>Enter examples (optional):</div>{{end}}
{{range $k, $v := .Tests}}
<div><input tabindex=2{{$k}} type=text name=t{{$k}} size=100 maxlength=1000 value="{{$v}}"></div>{{end}}
<div><input tabindex=99 type=submit value=Submit></div></form>
`))
