//  util.go -- miscellaneous utility helpers

package rx

import (
	"bufio"
	"log"
	"os"
	"strconv"
)

func Escape(c byte) string { // interpret meaning of \c
	// #%#%#% works only for escapes that repr a set of characters, not \b
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

func Protect(s string) string {  // protect unprintables with backslashes
        s = strconv.Quote(s)
	return s[1:len(s)-1]
}

func Cprotect(r rune) string {	// protect single unprintable char w/ backslash
	if strconv.IsPrint(r) {
		return string(r)
	} else {
		s := strconv.QuoteRune(r)
		return s[1:len(s)-1]
	}
}

func MkScanner(fname string) *bufio.Scanner { // return scanner for file
	if fname == "-" {
		return bufio.NewScanner(os.Stdin)
	} else {
		f, err := os.Open(fname)
		CkErr(err)
		return bufio.NewScanner(f)
	}
}

func CkErr(e error) { // abort if e is not nil
	if e != nil {
		log.Fatal(e)
	}
}
