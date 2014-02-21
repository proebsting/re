//  parse.go -- rx parse tree construction

//  Rx provides facilities for dealing with regular expressions.
package rx

import (
	"fmt"
	"regexp"
	"strconv"
)

// oprstack and exprstack are stacks that move in synchrony
var oprstack []byte  // operators associated with pushed expressions
var exprstack []Node // stack of pushed expressions

//  Parse parses a regular expression and returns a return tree of Nodes.
//  If there is an error, it returns nil (and an error).
//
//  Parse implements these forms:
//	abc  a|b|c  a(b|c)d
//	a?  b*  c+  d{m,n}
//	.  \d  \s  \w  [...]
func Parse(rexpr string) (Node, error) {

	var curr Node         // current parse tree
	var lside, rside Node // left and right side subtrees
	var cset *Cset        // computed character set

	curr = Epsilon()            // initialize empty parse tree
	oprstack = make([]byte, 0)  // initialize empty operator stack
	exprstack = make([]Node, 0) // initialize empty expresion stack
	orgstr := rexpr             // save original string

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

		case '|', '(':
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
					return nil, ParseError{orgstr, rexpr}
				}
				curr = Concatenate(lside, rside)
				continue
			}
			// no preceding '('!
			return nil, ParseError{orgstr, "unmatched ')'"}

		case '[':
			// bracket expression
			cset, rexpr = Bracketx(rexpr)
			if cset == nil {
				return nil, ParseError{orgstr, rexpr}
			}
			rside = MatchNode{cset}

		case '.': //#%#%#% no chars above 0x7F; this is a bug
			// wild character
			cset, _ = Bracketx("\x01-\x7F]")
			rside = MatchNode{cset}

		case '\\':
			// escaped character
			if len(rexpr) > 0 {
				cset, rexpr = Escape(rexpr)
				if cset == nil {
					return nil, ParseError{orgstr, rexpr}
				}
				rside = MatchNode{cset}
			} else {
				return nil, ParseError{orgstr, "'\\' at end"}
			}
		default:
			// single literal character
			rside = MatchNode{NewCset(string(thischar))}
		}

		// common code for handling postfix replication
		rside, rexpr = replicate(rside, rexpr)
		if rside == nil {
			return nil, ParseError{orgstr, rexpr}
		}
		curr = Concatenate(curr, rside)
	}

	curr = popAlts(curr) // check unpopped alternatives at end of string
	if len(oprstack) > 0 {
		return nil, ParseError{orgstr, "unclosed '('"}
	}
	return curr, nil
}

//  pops all consecutive alternatives from the operator/expression stacks
//  and returns the resulting subtree
func popAlts(d Node) Node {
	for j := len(oprstack) - 1; j >= 0 && oprstack[j] == '|'; j-- {
		d = Alternate(exprstack[j], d)
		oprstack = oprstack[0:j]
		exprstack = exprstack[0:j]
	}
	return d
}

var replx = regexp.MustCompile("{(\\d*)(,?)(\\d*)}") // expr for {n,m}

//  Replicate wraps a replication node around a subtree if it is followed by
//  posfix ?, *, +, or {m,n}.  It normally returns the resulting tree and
//  remaining string (both unmodified in the absence of postfix replication).
//  If there is an error in {m,n}, Replicate returns (nil, errmsg).
func replicate(d Node, p string) (Node, string) {
	if len(p) == 0 {
		return d, p
	}
	switch p[0] {
	case '?':
		return ReplNode{0, 1, d}, p[1:]
	case '*':
		return ReplNode{0, -1, d}, p[1:]
	case '+':
		return ReplNode{1, -1, d}, p[1:]
	case '{':
		result := replx.FindStringSubmatch(p)
		if result == nil {
			return nil, "malformed '{m,n}'"
		}
		p = p[len(result[0]):] // remove matched pattern
		minrep, err1 := strconv.Atoi(result[1])
		maxrep, err3 := strconv.Atoi(result[3])
		if err1 == nil && err3 == nil && minrep <= maxrep { // {n,m}
			return ReplNode{minrep, maxrep, d}, p
		}
		if err1 == nil && len(result[3]) == 0 {
			if len(result[2]) == 1 { // {n,}
				return ReplNode{minrep, -1, d}, p
			} else { // {n}
				return ReplNode{minrep, minrep, d}, p
			}
		}
		return nil, "malformed '{m,n}'"

	default:
		return d, p
	}
}

// ParseError diagnoses a malformed regular expression
type ParseError struct {
	BadExpr string
	Message string
}

func (e ParseError) Error() string {
	return fmt.Sprintf("rx: %s: in \"%s\"", e.Message, e.BadExpr)
}
