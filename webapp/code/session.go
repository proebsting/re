// session.go -- simple session management using cookies
//
// (so far this is just a proof of concept...)

package webapp

import (
	"net/http"
	"regexp"
	"time"
)

var ckpat = regexp.MustCompile(`^\d{8}$`) // cookie validation pattern

func initSession(w http.ResponseWriter, r *http.Request) {
	// define a cookie value based on current clock
	ckval := time.Now().Format("02150405")
	// always reset the Ephemeral cookie
	http.SetCookie(w, &http.Cookie{Name: "Ephemeral", Value: ckval})
	// set Session and Brief cookies if not already set
	cookie(w, r, "Session", ckval, time.Time{})
	cookie(w, r, "Brief", ckval, time.Now().Add(time.Minute))
}

func cookie(w http.ResponseWriter, r *http.Request,
	name string, value string, expires time.Time) string {
	c, _ := r.Cookie(name)
	if c != nil && ckpat.MatchString(c.Value) {
		return c.Value
	} else {
		http.SetCookie(w, &http.Cookie{
			Name: name, Value: value, Expires: expires})
		return value
	}
}
