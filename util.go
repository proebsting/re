//  util.go -- miscellaneous utility helpers

package rx

import (
	"bufio"
	"log"
	"os"
)

func Escape(c byte) byte { // interpret meaning of \c
	switch c {
		// #%#%#% need to flesh this out
		case 'e':	return '\033'
		case 't':	return '\t'
		case 'n':	return '\n'
		default:	return c
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
