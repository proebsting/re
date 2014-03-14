//  util.go -- miscellaneous utility helpers

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

//  Protect adds backslash notation, but no quotes, to protect unprintables.
func Protect(s string) string {
	s = strconv.Quote(s)
	return s[1 : len(s)-1]
}

//  Cprotect returns its argument if printable, else a backslash form.
func Cprotect(r rune) string {
	if strconv.IsPrint(r) {
		return string(r)
	} else {
		s := strconv.QuoteRune(r)
		return s[1 : len(s)-1]
	}
}

//  MkScanner creates a scanner for reading from a file, aborting on error.
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

// Jlist writes a slice to a file in JSON format, one entry per line.
// No newline is written at the end.
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
