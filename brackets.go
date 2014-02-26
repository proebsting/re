//  brac.go -- parsing of bracket expression

package rx

import (
	"fmt"
	"unicode"
)

var _ = fmt.Printf //#%#% for debugging

//  Bracketx parses a string as a bracket expression, returning the
//  computed set of characters and the remaining unprocessed part of s.
//  It assumes the introductory '[' has already been stripped from s.
//
//  If an error is found, bracketx returns (nil, errmsg).
//
//  Implements:  [abc] [^abc] [a-c] [\x]
func bracketx(s string) (*Cset, string) {

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
		case '[':
			// ordinary, but diagnose [:class:]
			if len(s) > 2 && s[0] == ':' &&
				unicode.IsLetter(rune(s[1])) {
				return nil, "[:class:] unimplemented"
			} else {
				result.Set(ch)
			}
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
				var eset *Cset
				eset, s = bescape(s)
				if eset == nil {
					return nil, s
				}
				result = result.Or(eset)
			} // else: error caught on next iteration
		default:
			// an ordinary char; add to set
			result.Set(ch)
		}
		cprev = ch
	}
	return nil, "unclosed '['"
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

//  Bescape interprets a backslash sequence in the context of a bracket
//  expression from which the initial \ has already been consumed.
//  In this context \b is a backspace.  Bescape returns the computed
//  cset and the remaining unescaped portion of the string.
//  If an error is found, bescape returns (nil, errmsg).
func bescape(s string) (*Cset, string) {
	if len(s) == 0 {
		return nil, "'\\' at end"
	}
	c := s[0]
	s = s[1:]
	switch c {
	case '0', '1', '2', '3', '4', '5', '6', '7':
		return nil, "'\\0nn' unimplemented"
	case 'a':
		return (&Cset{}).Set('\a'), s
	case 'b':
		return (&Cset{}).Set('\b'), s
	case 'c':
		return nil, "'\\cx' unimplemented"
	case 'd':
		return dset, s
	case 'e':
		return (&Cset{}).Set('\033'), s
	case 'f':
		return (&Cset{}).Set('\f'), s
	case 'n':
		return (&Cset{}).Set('\n'), s
	case 'p':
		return nil, "'\\px' unimplemented"
	case 'r':
		return (&Cset{}).Set('\r'), s
	case 's':
		return sset, s
	case 't':
		return (&Cset{}).Set('\t'), s
	case 'u':
		return nil, "'\\uhhhh' unimplemented"
	case 'v':
		return (&Cset{}).Set('\v'), s
	case 'w':
		return wset, s
	case 'x':
		return nil, "'\\xhh' unimplemented"
	case 'D':
		return dcompl, s
	case 'P':
		return nil, "'\\Px' unimplemented"
	case 'S':
		return scompl, s
	case 'W':
		return wcompl, s

	default:
		return (&Cset{}).Set(uint(c)), s
	}
}
