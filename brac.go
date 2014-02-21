//  brac.go -- parsing of bracket expression

package rx

import (
	"fmt"
)

var _ = fmt.Printf //#%#% for debugging

//  Bracketx parses a string as a bracket expression, returning the
//  computed set of characters and the remaining unprocessed part of s.
//  It assumes the introductory '[' has already been stripped from s.
//
//  Implements:  [abc] [^abc] [a-c] [\x]
func Bracketx(s string) (*Cset, string) {

	result := &Cset{}
	compl := false
	// check for initial '^'
	if len(s) > 0 && s[0] == '^' {
		compl = true
		s = s[1:]
	}
	cprev := uint(0) // no previous character
	// process body of expression
	for len(s) > 0 {
		ch := uint(s[0])
		s = s[1:]
		switch ch {
		case '-':
			// range of chars
			if cprev != 0 && len(s) > 0 && s[0] != ']' {
				ch = uint(s[0])
				s = s[1:]
				for j := cprev; j <= ch; j++ {
					result.Set(uint(j))
				}
			} else {
				result.Set(ch)
			}
		case ']':
			// set is complete unless this is first char
			if cprev != 0 {
				if compl {
					result = result.Compl()
				}
				return result, s
			} else {
				// initial ']' is ordinary
				result.Set(ch)
			}
		case '\\':
			if len(s) > 0 {
				result = result.Or(Escape(s[0]))
				s = s[1:]
			} // else: error caught on next iteration
		default:
			// an ordinary char; add to set
			result.Set(ch)
		}
		cprev = ch
	}
	// #%#% ERROR: no terminating ']'
	// for now, just treat as "[][]" (a cset of two chars ']' and '[')
	return NewCset("[]"), s
}

var dset, sset, wset, dcompl, scompl, wcompl *Cset

func init() {
	dset = NewCset("0123456789")
	sset = NewCset("\t\n\v\f\r ")
	wset = NewCset("0123456789_ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz")
	dcompl = dset.Compl()
	scompl = sset.Compl()
	wcompl = wset.Compl()

}

//  Escape returns the set of characters to be matched by \c inside
//  a bracket expression.  In this context \b is a backspace.
func Escape(c byte) *Cset {
	switch c {
	case 'a':
		return (&Cset{}).Set('\a')
	case 'b':
		return (&Cset{}).Set('\b')
	case 'd':
		return dset
	case 'e':
		return (&Cset{}).Set('\033')
	case 'f':
		return (&Cset{}).Set('\f')
	case 'n':
		return (&Cset{}).Set('\n')
	case 'r':
		return (&Cset{}).Set('\r')
	case 's':
		return sset
	case 't':
		return (&Cset{}).Set('\t')
	case 'v':
		return (&Cset{}).Set('\v')
	case 'w':
		return wset
	case 'D':
		return dcompl
	case 'S':
		return scompl
	case 'W':
		return wcompl

	default:
		return (&Cset{}).Set(uint(c))
	}
}
