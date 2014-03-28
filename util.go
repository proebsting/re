//  util.go -- miscellaneous utility helpers
//
//  This file collects small and often unrelated general-purpose helper
//  functions that are not closely associated with any other file.

package rx

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"reflect"
	"strconv"
	"syscall"
	"time"
)

//  CkErr aborts with a fatal error if e is not nil.
func CkErr(e error) { // abort if e is not nil
	if e != nil {
		log.Fatal(e)
	}
}

//  Protect adds backslash notation, but no quotes,
//  to protect unprintables in a string.
func Protect(s string) string {
	s = strconv.Quote(s)
	return s[1 : len(s)-1]
}

//  CPUtime returns the current CPU usage (user time + system time).
func CPUtime() time.Duration {
	var ustruct syscall.Rusage
	CkErr(syscall.Getrusage(0, &ustruct))
	user := time.Duration(syscall.TimevalToNsec(ustruct.Utime))
	sys := time.Duration(syscall.TimevalToNsec(ustruct.Stime))
	return user + sys
}

//  Interval returns the CPU time (user + system) since the preceding call.
func Interval() time.Duration {
	total := CPUtime()
	delta := total - prevTotal
	prevTotal = total
	return delta
}

var prevTotal time.Duration // total time at list check

//  ShowInterval calcs and (unless label is empty) prints the last interval.
func ShowInterval(label string) {
	dt := Interval().Seconds()
	if label != "" {
		fmt.Printf("%7.3f %s\n", dt, label)
	}
}

//  Jlist writes a slice of anything Marshalable to a file in JSON format,
//  one entry per line.  No newline is written at the end.
func Jlist(f io.Writer, slc interface{}) {
	switch reflect.TypeOf(slc).Kind() {
	case reflect.Slice:
		a := reflect.ValueOf(slc)
		n := a.Len()
		fmt.Fprintln(f, "[")
		for i := 0; i < n; i++ {
			json, err := json.Marshal(a.Index(i).Interface())
			CkErr(err)
			if i < n-1 {
				json = append(json, ',')
			}
			fmt.Fprintf(f, "%s\n", string(json))
		}
		fmt.Fprint(f, "]")
	default:
		panic("Jlist: unimplemented type")
	}
}
