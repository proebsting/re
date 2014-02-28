//  bracets.go -- parsing of bracket expression

package rx

import (
	"fmt"
	"unicode"
)

var _ = fmt.Printf //#%#% for debugging

//  Bxparse parses a string as a bracket expression, returning the
//  computed set of characters and the remaining unprocessed part of s.
//  It assumes the introductory '[' has already been stripped from s.
//
//  If an error is found, bxparse returns (nil, errmsg).
//
//  Implements:  [abc] [^abc] [a-c] [\x]
func bxparse(s string) (*BitSet, string) {

	result := &BitSet{}
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
					result = result.CharCompl()
				}
				return result, s
			} else {
				// initial ']' is ordinary
				result.Set(ch)
			}
		case '\\':
			if len(s) > 0 {
				var eset *BitSet
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

var dset, sset, wset, dcompl, scompl, wcompl *BitSet

func init() {
	dset = CharSet("0123456789")
	sset = CharSet("\t\n\v\f\r ")
	wset = CharSet("0123456789_ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz")
	dcompl = dset.CharCompl()
	scompl = sset.CharCompl()
	wcompl = wset.CharCompl()

}

//  Bescape interprets a backslash sequence in the context of a bracket
//  expression from which the initial \ has already been consumed.
//  In this context \b is a backspace.  Bescape returns the computed
//  cset and the remaining unescaped portion of the string.
//  If an error is found, bescape returns (nil, errmsg).
func bescape(s string) (*BitSet, string) {
	if len(s) == 0 {
		return nil, "'\\' at end"
	}
	c := s[0]
	s = s[1:]
	switch c {
	case '0', '1', '2', '3', '4', '5', '6', '7':
		return nil, "'\\0nn' unimplemented"
	case 'a':
		return (&BitSet{}).Set('\a'), s
	case 'b':
		return (&BitSet{}).Set('\b'), s
	case 'c':
		return nil, "'\\cx' unimplemented"
	case 'd':
		return dset, s
	case 'e':
		return (&BitSet{}).Set('\033'), s
	case 'f':
		return (&BitSet{}).Set('\f'), s
	case 'n':
		return (&BitSet{}).Set('\n'), s
	case 'p':
		return nil, "'\\px' unimplemented"
	case 'r':
		return (&BitSet{}).Set('\r'), s
	case 's':
		return sset, s
	case 't':
		return (&BitSet{}).Set('\t'), s
	case 'u':
		return nil, "'\\uhhhh' unimplemented"
	case 'v':
		return (&BitSet{}).Set('\v'), s
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
		return (&BitSet{}).Set(uint(c)), s
	}
}
