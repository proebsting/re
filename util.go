//  util.go -- miscellaneous utility helpers
//
//  This file collects small and often unrelated general-purpose helper
//  functions that are not closely associated with any other file.

package rx

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"reflect"
	"strconv"
)

//  MkScanner creates a Scanner for reading from a file, aborting on error.
//  A filename of "-" reads standard input.
func MkScanner(fname string) *bufio.Scanner { // return scanner for file
	if fname == "-" {
		return bufio.NewScanner(os.Stdin)
	} else {
		f, err := os.Open(fname)
		CkErr(err)
		return bufio.NewScanner(f)
	}
}

//  CkErr aborts with a fatal error if e is not nil.
func CkErr(e error) { // abort if e is not nil
	if e != nil {
		log.Fatal(e)
	}
}

//  IsComment returns true if a line begins with '#' or is empty.
func IsComment(s string) bool {
	return len(s) == 0 || s[0] == '#'
}

//  Protect adds backslash notation, but no quotes,
//  to protect unprintables in a string.
func Protect(s string) string {
	s = strconv.Quote(s)
	return s[1 : len(s)-1]
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
