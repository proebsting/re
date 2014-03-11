//  util.go -- miscellaneous utility helpers

package rx

import (
	"bufio"
	"log"
	"os"
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

//  IsComment returns true if a line begins with '#'.
func IsComment(s string) bool {
	return len(s) > 0 && s[0] == '#'
}
