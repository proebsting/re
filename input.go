//  input.go -- regular expression input

package rx

import (
	"bufio"
	"flag"
	"log"
	"os"
)

//  A RegExParsed is a single parsed regular expression.
//  If Tree is not nil then the expression was parsed as valid.
//  If Tree is nil and Err is not, Err is a parsing error.
//  If Tree is nil and Err is nil, the struct can represent a comment.
type RegExParsed struct {
	Expr string // input string
	Tree Node   // parse tree
	Err  error  // parse error
	//#%#% to come: add metadata field(s)
}

//  RegExParsed.IsExpr returns true if the struct represents an expression,
//  not a comment, whether or not it is valid.
func (rxp *RegExParsed) IsExpr() bool {
	return rxp.Tree != nil || rxp.Err != nil
}

//  LoadExpressions reads a file and parses the expressions found.
//  A filename of "" or "-" reads from standard input.  Any file error is fatal.
//
//  Empty lines and lines beginning with '#' are treated as comments.
//  If non-nil, the function f is called for each line read.
//  The returned array contains only successfully parsed expressions.
func LoadExpressions(fname string, f func(*RegExParsed)) []*RegExParsed {
	efile := MkScanner(fname)
	elist := make([]*RegExParsed, 0)
	for efile.Scan() {
		line := efile.Text()
		e := &RegExParsed{Expr: line}
		if !IsComment(line) {
			e.Tree, e.Err = Parse(line)
			if e.Tree != nil {
				elist = append(elist, e)
			}
		}
		if f != nil {
			f(e)
		}
	}
	CkErr(efile.Err())
	return elist
}

//  OneInputFile returns the name of the input file from the command line.
//  The program is aborted with an error message if multiple arguments appear.
//  If no arguments are present, "-" is returned to represent standard input.
func OneInputFile() string {
	flag.Parse() // in case not already called
	args := flag.Args()
	switch len(args) {
	case 0:
		return "-"
	case 1:
		return args[0]
	default:
		log.Fatal("too many arguments")
	}
	return "" //NOTREACHED
}

//  MkScanner creates a Scanner for reading from a file, aborting on error.
//  A filename of "-" reads standard input.
func MkScanner(fname string) *bufio.Scanner { // return scanner for file
	if fname == "" || fname == "-" {
		return bufio.NewScanner(os.Stdin)
	} else {
		f, err := os.Open(fname)
		CkErr(err)
		return bufio.NewScanner(f)
	}
}

//  IsComment returns true if a line begins with '#' or is empty.
func IsComment(s string) bool {
	return len(s) == 0 || s[0] == '#'
}
