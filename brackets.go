//  brackets.go -- parsing of bracket expression

package rx

import (
	"fmt"
	"strconv"
	"unicode"
)

//  bxparse parses a string as a bracket expression, returning the
//  computed set of characters and the remaining unprocessed part of s.
//  It assumes the introductory '[' has already been stripped from s.
//
//  If an error is found, bxparse returns (nil, errmsg).
//
//  bxparse implements:  [abc] [^abc] [a-c] [\x]
func bxparse(s []rune) (*BitSet, []rune) {

	result := &BitSet{}
	compl := false

	// check for initial '^'
	if len(s) > 0 && s[0] == '^' {
		compl = true
		s = s[1:]
	}
	cprev := 0 // no previous character
	// process body of expression
	for len(s) > 0 {
		ch := int(s[0])
		s = s[1:]
		switch ch {
		case '[':
			// ordinary, but diagnose [:class:]
			if len(s) > 2 && s[0] == ':' &&
				unicode.IsLetter(s[1]) {
				return nil, []rune("[:class:] unimplemented")
			} else {
				result.Set(ch)
			}
		case '-':
			// range of chars
			if cprev != 0 && len(s) > 0 && s[0] != ']' {
				ch = int(s[0])
				s = s[1:]
				if ch == '\\' {
					var eset *BitSet
					eset, s = bescape(s)
					if eset == nil {
						return nil, s
					}
					ch = eset.LowBit()
				}
				if ch < cprev {
					return nil, []rune("invalid range")
				}
				for j := cprev; j <= ch; j++ {
					result.Set(j)
				}
			} else {
				result.Set(ch)
			}
		case ']':
			// set is complete unless this is first char
			if !result.IsEmpty() {
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
				result.OrWith(eset)
				ch = eset.HighBit()
			} // else: error caught on next iteration
		default:
			// an ordinary char; add to set
			result.Set(ch)
		}
		cprev = ch
	}
	return nil, []rune("unclosed '['")
}

//  bescape interprets a backslash sequence in the context of a bracket
//  expression from which the initial \ has already been consumed.
//  In this context \b is a backspace.  bescape returns the computed
//  charset and the remaining unescaped portion of the string.
//  If an error is found, bescape returns (nil, errmsg).
//
//  bescape implements:
//	\a \b \e \f \n \r \t \v \046 \xF7 \u03A8
//	\d \s \w \D \S \W
func bescape(s []rune) (*BitSet, []rune) {
	if len(s) == 0 {
		return nil, []rune("'\\' at end")
	}
	c := int(s[0])
	s = s[1:]
	switch c {
	case '0', '1', '2', '3', '4', '5', '6', '7':
		v := c - '0'               // first digit
		if o := octal(s); o >= 0 { // optional 2nd digit
			v = 8*v + o
			s = s[1:]
		}
		if o := octal(s); o >= 0 { // optional 3nd digit
			v = 8*v + o
			s = s[1:]
		}
		return (&BitSet{}).Set(v), s
	case 'a':
		return (&BitSet{}).Set('\a'), s
	case 'b':
		return (&BitSet{}).Set('\b'), s
	case 'c':
		return nil, []rune("'\\cx' unimplemented")
	case 'd':
		return DigitSet, s
	case 'e':
		return (&BitSet{}).Set('\033'), s
	case 'f':
		return (&BitSet{}).Set('\f'), s
	case 'n':
		return (&BitSet{}).Set('\n'), s
	case 'p':
		return nil, []rune("'\\px' unimplemented")
	case 'r':
		return (&BitSet{}).Set('\r'), s
	case 's':
		return SpaceSet, s
	case 't':
		return (&BitSet{}).Set('\t'), s
	case 'u':
		v := hexl(s, 4)
		if v >= 0 {
			return (&BitSet{}).Set(v), s[4:]
		} else {
			return nil, []rune("malformed '\\uhhhh'")
		}
	case 'v':
		return (&BitSet{}).Set('\v'), s
	case 'w':
		return WordSet, s
	case 'x':
		v := hexl(s, 2)
		if v >= 0 {
			return (&BitSet{}).Set(v), s[2:]
		} else {
			return nil, []rune("malformed '\\xhh'")
		}
	case 'D':
		return NonDigit, s
	case 'P':
		return nil, []rune("'\\Px' unimplemented")
	case 'S':
		return NonSpace, s
	case 'W':
		return NonWord, s

	default:
		if unicode.IsLetter(rune(c)) {
			return nil, []rune(fmt.Sprintf("'\\%c' unrecognized", c))
		} else {
			return (&BitSet{}).Set(c), s
		}
	}
}

//  octal returns the value of the first digit of s, or -1 if not octal digit.
func octal(s []rune) int {
	if len(s) > 0 && s[0] >= '0' && s[0] <= '7' {
		return int(s[0]) - '0'
	} else {
		return -1
	}
}

//  hexl returns the value of the first n hex digits of s, or -1 if bad.
func hexl(s []rune, n int) int {
	if len(s) < n {
		return -1
	}
	v, err := strconv.ParseInt(string(s[0:n]), 16, 64)
	if err == nil {
		return int(v)
	} else {
		return -1
	}
}
