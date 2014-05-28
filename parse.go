//  parse.go -- regular expression scanning and parse tree construction

package rx

import (
	"fmt"
	"regexp"
	"strconv"
)

//  oprstack and exprstack are stacks that move in synchrony.
var oprstack []rune  // operators associated with pushed expressions
var exprstack []Node // stack of pushed expressions

//  Parse parses a regular expression and returns a tree of Nodes.
//  If there is an error, it returns nil (and an error).
//
//  Parse implements these forms:
//	abc  a|b|c  a(b|c)d
//	a?  b*  c+  d{m,n}
//	\a \e \f \n \r \t \v \046 \xF7 \u03A8
//	.  \d \s \w \D \S \W
//	[abc]  [^abc]  [a-c]  [\x]
//
//  Parse ignores the Perl non-capturing submatch form "(?:",  but other
//  "(?" forms are errors.
//
//  All trees are "anchored".  An initial '^' and/or final '$' is ignored.
//
//  Wildcard character sets (for ".", "\w", "\D", "[%\d]", etc.)
//  are limited to the ASCII character set [\x01-\x7F].
func Parse(orgstr string) (Node, error) {

	var curr Node         // current parse tree
	var lside, rside Node // left and right side subtrees
	var cset *BitSet      // computed character set

	curr = Epsilon()            // initialize empty parse tree
	oprstack = make([]rune, 0)  // initialize empty operator stack
	exprstack = make([]Node, 0) // initialize empty expression stack

	// convert UTF-8 string into slice of runes
	rexpr := []rune(orgstr)

	if len(rexpr) > 0 && rexpr[0] == '^' {
		rexpr = rexpr[1:] // remove initial '^'
	}

	for len(rexpr) > 0 { // for every character in regexp
		// invariant: curr holds the parse tree completed so far
		// invariant: rexpr holds remaining unprocessed characters

		// dispatch based on next character
		thischar := rexpr[0] // get character
		rexpr = rexpr[1:]    // trim from string
		switch thischar {    // dispatch

		case '?', '*', '+':
			// ignore:  if seen here, any of these characters
			// just replicates an empty string to no effect
			continue

		case '^':
			// '^' is illegal except (handled earlier) when first
			return nil, &ParseError{
				orgstr, "Embedded '^' unimplemented"}

		case '$':
			// '^' is illegal except when last (then ignore)
			if len(rexpr) != 0 {
				return nil, &ParseError{
					orgstr, "Embedded '$' unimplemented"}
			}
			continue

		case '(':
			// check for "(?..." forms
			// treat "(?:" as "(" but abort on other "(?...")
			if len(rexpr) > 1 && rexpr[0] == '?' {
				if rexpr[1] == ':' {
					// just ignore "?:"
					rexpr = rexpr[2:]
				} else {
					return nil, &ParseError{
						orgstr, "'(?...' unimplemented"}
				}
			}
			fallthrough // the rest is common with '|'
		case '|': // and '(' falling through
			// swap of context: clear current expression after
			// pushing it on a stack (and push opr on another)
			oprstack = append(oprstack, thischar)
			exprstack = append(exprstack, curr)
			curr = Epsilon()
			continue // don't check/allow postfix replication

		case ')':
			// close parenthesis: gather parts together
			curr = popAlts(curr)   // first handle alternation
			j := len(oprstack) - 1 // pop the opening paren
			if j >= 0 && oprstack[j] == '(' {
				lside = exprstack[j]       // predecessor
				exprstack = exprstack[0:j] // pop stack
				oprstack = oprstack[0:j]   // pop opr
				rside, rexpr = replicate(curr, rexpr)
				if rside == nil {
					return nil, RuneError(orgstr, rexpr)
				}
				curr = Concatenate(lside, rside)
				continue
			}
			// no preceding '('!
			return nil, &ParseError{orgstr, "unmatched ')'"}

		case '[':
			// bracket expression
			cset, rexpr = bxparse(rexpr)
			if cset == nil {
				return nil, RuneError(orgstr, rexpr)
			}
			rside = MatchAny(cset)

		case '.':
			// wild character
			cset = AllChars
			rside = MatchAny(cset)

		case '\\':
			rside, rexpr = rescape(rexpr)
			if rside == nil {
				return nil, RuneError(orgstr, rexpr)
			}
		default:
			// single literal character
			rside = MatchAny(CharSet(string(thischar)))
		}

		// common code for handling postfix replication
		rside, rexpr = replicate(rside, rexpr)
		if rside == nil {
			return nil, RuneError(orgstr, rexpr)
		}
		curr = Concatenate(curr, rside)
	}

	curr = popAlts(curr) // check unpopped alternatives at end of string
	if len(oprstack) > 0 {
		return nil, &ParseError{orgstr, "unclosed '('"}
	}
	// success!
	return curr, nil // return parse tree
}

//  popAlts pops all consecutive alternatives from the operator/expression
//  stacks and returns the resulting subtree.
func popAlts(d Node) Node {
	for j := len(oprstack) - 1; j >= 0 && oprstack[j] == '|'; j-- {
		d = Alternate(exprstack[j], d)
		oprstack = oprstack[0:j]
		exprstack = exprstack[0:j]
	}
	return d
}

//  Four helper functions consume regexp characters:
//  	replicate(), rescape(), bxparse(), and bescape().
//  These functions obey the following conventions.
//
//  Each helper function accepts the current remaining portion of the regexp
//  as a []rune parameter (possibly not the only parameter) and returns two
//  values.  The first is the "primary" function return, for example a Node.
//  The second is the updated slice of remaining characters.
//
//  If an error is detected, the primary return is nil, and the second return
//  is an error message as an array of runes.

var replx = regexp.MustCompile("{(\\d*)(,?)(\\d*)}") // expr for {m,n}

//  replicate wraps a replication node around a subtree if it is followed by
//  posfix ?, *, +, or {m,n}.  It normally returns the resulting tree and
//  remaining string (both unmodified in the absence of postfix replication).
//  If there is an error in {m,n}, replicate returns (nil, errmsg).
func replicate(d Node, p []rune) (Node, []rune) {
	if len(p) == 0 {
		return d, p
	}
	switch p[0] {
	case '?':
		return replval(0, 1, d, p[1:])
	case '*':
		return replval(0, -1, d, p[1:])
	case '+':
		return replval(1, -1, d, p[1:])
	case '{':
		result := replx.FindStringSubmatch(string(p))
		if result == nil {
			return nil, []rune(`malformed "{m,n}"`)
		}
		p = p[len(result[0]):] // remove matched pattern
		minrep, err1 := strconv.Atoi(result[1])
		maxrep, err3 := strconv.Atoi(result[3])
		if len(result[3]) == 0 {
			if len(result[2]) == 1 {
				maxrep, err3 = -1, nil // {m,}
			} else {
				maxrep, err3 = minrep, nil // {m} means {m,m}
			}
		}
		if err1 != nil || err3 != nil {
			return nil, []rune(`malformed "{m,n}"`)
		}
		if maxrep < 0 || maxrep > minrep { // {m,} or {m,n}
			return replval(minrep, maxrep, d, p)
		}
		if maxrep < minrep {
			return nil, []rune(`malformed "{m,n}"`)
		}
		// now we have maxrep == minrep and valid
		if maxrep == 0 {
			return Epsilon(), p // {0} or {0,0}
		} else if maxrep == 1 {
			return d, p // {1} or {1,1}
		} else {
			return replval(minrep, maxrep, d, p) // expand later
		}

	default:
		return d, p
	}
}

//  replval constructs the return value for replicate and checks for
//  an illegal second replication operator or "prefer-fewer" '?'
func replval(min int, max int, d Node, remdr []rune) (Node, []rune) {
	if len(remdr) > 0 {
		switch remdr[0] {
		case '?':
			return nil, []rune("prefer-fewer '?' unimplemented")
		case '*', '+', '{':
			return nil, []rune("multiple adjacent duplication symbols")
		}
	}
	return &ReplNode{min, max, d, nildata}, remdr
}

//  rescape handles a backslash encountered at the regexp level
//  (not inside a bracket expression).  It assumes the backslash
//  has already consumed, and returns the resulting Node and the
//  updated string after processing the escaped characters.
//  In case of error it returns (nil, errmsg).
func rescape(rexpr []rune) (Node, []rune) {
	if len(rexpr) == 0 {
		return nil, []rune("'\\' at end")
	}
	ch := rexpr[0]
	switch ch {
	case 'b':
		return nil, []rune("\\b (boundary) unimplemented")
	case '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return nil, []rune(fmt.Sprintf("\\%c (backref) unimplemented", ch))
	default:
		// otherwise meaning is the same as in a bracket expression
		cset, rexpr := bescape(rexpr)
		if cset == nil {
			return nil, rexpr
		}
		return MatchAny(cset), rexpr
	}
}

//  ParseError diagnoses a malformed regular expression.
type ParseError struct {
	BadExpr string
	Message string
}

//  ParseError.Error formats a parser error for printing.
func (e ParseError) Error() string {
	return fmt.Sprintf("rx: %s: in \"%s\"", e.Message, e.BadExpr)
}

//  RuneError constructs a ParseError from a string and a message in runes
func RuneError(badexpr string, message []rune) *ParseError {
	return &ParseError{badexpr, string(message)}
}
