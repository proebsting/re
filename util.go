//  util.go -- miscellaneous utility helpers

package rx

import (
	"bufio"
	"log"
	"os"
	"strconv"
)

//  Escape returns the set of characters to be matched by \c.
func Escape(c byte) string {
	//  #%#%#% works only for escapes that repr a set of characters, not \b
	switch c {
	case 'd':
		return "0123456789"
	case 's':
		return "\t\n\v\f\r "
	case 'w':
		return "0123456789_ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	default:
		return string(c)
	}
}

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
