//  charset.go -- bit set extensions for use as sets of characters
//
//  These additional functions support the use of a BitSet as a set of chars.
//  No distinct type is defined, however -- it's still a BitSet.
//
//  Note that "all characters" (for purposes of wildcarding or complementing)
//  defines a set of just the ASCII characters [\x01-\x0F].

package rx

import (
	"math/rand"
	"strconv"
)

//  Predefined global character sets.
//  Once constructed, these should be treated as constant.
var (
	SpaceSet  *BitSet = CharSet("\t\n\v\f\r ")
	DigitSet  *BitSet = CharSet("0123456789")
	UpperSet  *BitSet = CharSet("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	LowerSet  *BitSet = CharSet("abcdefghijklmnopqrstuvwxyz")
	LetterSet *BitSet = UpperSet.Or(LowerSet)
	WordSet   *BitSet = LetterSet.Or(DigitSet).Set('_')
	CtrlSet   *BitSet = CharRange('\x00', '\x1F').Or(CharRange('\x7F', '\x9F'))
	AllChars  *BitSet = CharRange('\x01', '\x7F') // matched by "."
	NonDigit  *BitSet = DigitSet.CharCompl()
	NonSpace  *BitSet = SpaceSet.CharCompl()
	NonWord   *BitSet = WordSet.CharCompl()
)

//  CharSet makes a BitSet from a string of member characters.
func CharSet(s string) *BitSet {
	cs := new(BitSet)
	for _, ch := range s {
		cs.Set(int(ch))
	}
	return cs
}

//  CharRange makes a BitSet from a range of characters.
func CharRange(low int, high int) *BitSet {
	cs := new(BitSet)
	for ; low <= high; low++ {
		cs.Set(low)
	}
	return cs
}

//  BitSet.CharCompl produces a new BitSet that is the complement of its inputs
//  with respect to the universe of matchable characters AllChars.
func (b1 *BitSet) CharCompl() *BitSet {
	b3 := new(BitSet)
	b3.Bits.Xor(&b1.Bits, &AllChars.Bits)
	return b3
}

//  BitSet.RandChar returns a single randomly chosen BitSet element.
//  Control characters are avoided unless nothing else is available.
func (b *BitSet) RandChar() rune {
	low := b.LowBit()            // lowest eligible char
	high := b.HighBit()          // highest eligible char
	if low < ' ' || high > '~' { // if range extends beyond ASCII printables
		b2 := b.AndNot(CtrlSet) // remove the control characters
		if !b2.IsEmpty() {      // and use that set if non-empty
			low = b2.LowBit()
			high = b2.HighBit()
		}
	}
	//  pick a random char between low and high inclusive.
	//  if it's part of the set, we're done.
	c := low + rand.Intn(high-low+1)
	if b.Test(c) {
		return rune(c)
	}
	//  otherwise, pick a random stride and find one.
	span := high - low + 1
	stride := rand.Intn(span)
	for GCD(stride, span) > 1 {
		stride--
	}
	c = low + ((c - low + stride) % span)
	for !b.Test(c) {
		c = low + ((c - low + stride) % span)
	}
	return rune(c)
}

//  BitSet.Bracketed() returns a bracket-expression form of a character set,
//  using ranges if appropriate and escaping (only) unprintables.
//  ] and - are specially placed at beginning and end respectively.
func (b *BitSet) Bracketed() string {
	l := b.LowBit()
	h := b.HighBit()
	open := "["
	close := "]"
	s := make([]byte, 0)
	for i := l; i <= h; i++ { // for all chars up to highest
		if b.Test(i) { // if char is included
			if i == ']' {
				open = "[]" // move ']' to front
				continue
			}
			if i == '-' {
				close = "-]" // defer '-' to end
				continue
			}
			s = append(s, cprotect(rune(i))...) // show char
			var j int
			for j = i + 1; b.Test(j); j++ {
				// count consecutive inclusions
			}
			if j-i > 3 { // if worth using [a-z] form
				if (j-1) == '-' || (j-1) == ']' {
					// don't end range with '-' or ']'
					j--
				}
				i = j - 1
				s = append(s, '-')
				s = append(s, cprotect(rune(i))...)
			}
		}
	}
	return open + string(s) + close
}

//  cprotect returns its argument if printable, else a backslash form.
func cprotect(r rune) string {
	if strconv.IsPrint(r) {
		return string(r)
	} else {
		s := strconv.QuoteRune(r)
		return s[1 : len(s)-1]
	}
}
