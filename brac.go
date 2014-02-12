//  brac.go -- parsing of bracket expression

package rx

import (
	"fmt"
)

//  bracketx(s) parses string s as a bracket expression,
//  returning the computed Cset and the remaining unprocessed string.

//#%#% incomplete; does not yet handle [^abc] or [\x] or [:digits:] etc

func bracketx(s string) (*Cset, string) {

	if len(s) == 0 || s[0] != '[' {
		panic(fmt.Sprintf("not a bracket expr: \"%s\"", s))
	}
	chars := make([]byte, 0)

	for i := 1; i < len(s); i++ {
		ch := s[i]
		switch (ch) {
			case '-':
				// range of chars
				if i > 1 && i+1 < len(s) && s[i+1] != ']' {
					i++;
					for j := s[i-2]; j <= s[i]; j++ {
						chars = append(chars, j)
					}
				} else {
					chars = append(chars, '-')
				}
			case ']':
				// set is complete unless this is first char
				if i > 1 {
					return NewCset(string(chars)), s[i+1:]
				} else {
					// initial ']' is ordinary
					chars = append(chars, s[i])
				}
			case '\\':
				if i+1 < len(s) {
					i++
					chars = append(chars, Escape(s[i]))
				} // else: error caught on next iteration
			default:
				// an ordinary char; add to set
				chars = append(chars, s[i])
		}
	}
	// #%#% ERROR: no terminating ']'
	// for now, just treat as '[[]'
	return NewCset(s[0:1]), s[1:]
}
