//  form.go -- form generation

package webapp

import (
	"html/template"
	"io"
)

//  putform outputs a form for submitting n expressions
func putform(w io.Writer, target string, label string, n int, values []string) {
	for len(values) < n {
		values = append(values, "")
	}
	tForm.Execute(w, struct {
		Target string
		Label  string
		Values []string
	}{
		target, label, values,
	})
}

var tForm = template.Must(template.New("form").Parse(`
<form action="{{.Target}}" method=post>
<div><div style="float:left;">{{.Label}} &nbsp;</div>
<div style="font-size: 67%;">
<button type=button class=link onclick="clearForm(this.form);">
(clear form)</button></div></div>
<div class=reset></div>
{{range $k, $v := .Values}}
<div><input tabindex=1{{$k}} type=text name=a{{$k}} size=100 maxlength=1000 value="{{$v}}"></div>{{end}}
<div><input tabindex=99 type=submit value=Submit></div></form>
`))
