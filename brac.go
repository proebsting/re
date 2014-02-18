//  brac.go -- parsing of bracket expression

package rx

import (
	"fmt"
)

//  bracketx parses a string as a bracket expression,
//  returning the computed Cset and the remaining unprocessed part of s.
//  It assumes the introductory '[' has already been stripped from s.
//  #%#% Incomplete; does not yet handle [^abc] or [:digits:] etc
func bracketx(s string) (*Cset, string) {

	chars := make([]byte, 0)
	for len(s) > 0 {
		ch := s[0]
		s = s[1:]
		switch ch {
		case '-':
			// range of chars
			if len(chars) > 0 && len(s) > 0 && s[0] != ']' {
				ch = s[0]
				s = s[1:]
				for j := chars[len(chars)-1]; j <= ch; j++ {
					chars = append(chars, j)
				}
			} else {
				chars = append(chars, '-')
			}
		case ']':
			// set is complete unless this is first char
			if len(chars) > 0 {
				return NewCset(string(chars)), s
			} else {
				// initial ']' is ordinary
				chars = append(chars, ']')
			}
		case '\\':
			if len(s) > 0 {
				chars = append(chars, Escape(s[0])...)
				s = s[1:]
			} // else: error caught on next iteration
		default:
			// an ordinary char; add to set
			chars = append(chars, ch)
		}
	}
	// #%#% ERROR: no terminating ']'
	// for now, just treat as "[][]" (a cset of two chars ']' and '[')
	return NewCset("][]"), s
}

func unwanted() { fmt.Println(00) } //#%#% just to ensure a fmt reference
