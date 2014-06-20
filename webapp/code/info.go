//  info.go -- generate info page (developer details)

package webapp

import (
	"appengine"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"runtime"
	"rx"
	"strings"
	"time"
)

//  info generates a page of information useful to the developers
func info(w http.ResponseWriter, r *http.Request) {

	// must do all reading before any writing
	var data struct {
		Req    *http.Request
		Body   []byte
		BE     error
		Dctr   string
		GoArch string
		GoOs   string
		GoVer  string
		VID    string
		Vtime  string
		MaxCx  int
	}
	data.Req = r
	data.Body, data.BE = ioutil.ReadAll(r.Body)
	data.Dctr = appengine.Datacenter()
	data.GoArch = runtime.GOARCH
	data.GoOs = runtime.GOOS
	data.GoVer = runtime.Version()
	data.VID = appengine.VersionID(appengine.NewContext(r))
	data.MaxCx = rx.MaxComplexity

	var ver int
	var bigtime int64
	fmt.Sscanf(data.VID, "%d.%d", &ver, &bigtime)
	if strings.HasPrefix(r.Host, "localhost:") {
		// don't know how to decode this in SDK environment
		data.Vtime = "?!"
	} else {
		// decode VID to match with appengine admin logs
		t := time.Unix(bigtime>>28, 0)
		data.Vtime = t.Format("01/02 15:04")
	}

	putheader(w, r, "Info")

	fmt.Fprint(w, "<P>")
	for i := 0; i < 10; i++ {
		fmt.Fprintf(w, "<span class=c%d>= c%d =</span>&nbsp;\n", i, i)
	}
	fmt.Fprintf(w, "<span class=cg>= cg =</span>&nbsp;\n") // .cg
	fmt.Fprintf(w, "<span class=cw>= cw =</span>&nbsp;\n") // .cw
	fmt.Fprintln(w)
	tInfo.Execute(w, data)
	putfooter(w, r)
}

var tInfo = template.Must(template.New("info").Parse(
	`<P>Host: {{.Req.Host}}
<BR>Datacenter: {{.Dctr}} ({{.GoArch}} {{.GoOs}})
<BR>Go Version: {{.GoVer}}
<BR>App Version ID: {{.VID}} ({{.Vtime}})
<BR><BR>MaxComplexity: {{.MaxCx}}
<H2>Request Header</H2>
<P>{{range $k, $v := .Req.Header}}{{$k}} : {{$v}}<BR>
{{end}}
<H2>Request Body</H2>
{{printf "%s" .Body}}
{{if .BE}}<P><B>BODY ERROR:</B> {{.BE}}{{end}}
`))
