//  util.go -- miscellaneous utility helpers

package rx

import (
	"bufio"
	"log"
	"os"
	"strconv"
)

//  Escape returns the set of characters to be matched by \c inside a
//  bracket expression.  This works for \d \s \w which represent small
//  finite sets, but not their complements \D \S \W.
//  In this context \b is a backspace and other C escapes are similar.
func Escape(c byte) string {
	switch c {
	case 'a':
		return "\a"
	case 'b':
		return "\b"
	case 'd':
		return "0123456789"
	case 'e':
		return "\033"
	case 'f':
		return "\f"
	case 'n':
		return "\n"
	case 'r':
		return "\r"
	case 's':
		return "\t\n\v\f\r "
	case 't':
		return "\t"
	case 'v':
		return "\v"
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
