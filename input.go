//  input.go -- regular expression input

package rx

import (
	"bufio"
	"flag"
	"log"
	"os"
	"regexp"
)

//  Globals set as a side effect of loading input
var InputRegExCount int // number of expressions successfully loaded
var InputErrorCount int // number of unacceptable expressions rejected

//  A RegExParsed is a single parsed regular expression.
//  If Tree is not nil then the expression was parsed as valid.
//  If Tree is nil and Err is not, Err is a parsing error.
//  If Tree is nil and Err is nil, the struct can represent a comment.
//  Metadata for a comment reflects current line only; otherwise cumulative.
type RegExParsed struct {
	Expr string            // input string
	Tree Node              // parse tree
	Err  error             // parse error
	Meta map[string]string // metadata list
}

//  RegExParsed.IsExpr returns true if the struct represents an expression,
//  not a comment, whether or not it is valid.
func (rxp *RegExParsed) IsExpr() bool {
	return rxp.Tree != nil || rxp.Err != nil
}

//  LoadExpressions reads a file and parses the expressions found.
//  A filename of "" or "-" reads from standard input.  Any file error is fatal.
//  See LoadFromScanner for details.
func LoadExpressions(fname string, f func(*RegExParsed)) []*RegExParsed {
	return LoadFromScanner(MkScanner(fname), f)
}

//  LoadFromScanner reads and parses expressions from a bufio.Scanner.
//
//  Empty lines and lines beginning with '#' are treated as comments.
//  If non-nil, the function f is called for each line read.
//  The returned array contains only successfully parsed expressions.
//
//  Metadata from comments matching the pattern "^#\w+:" is accumulated and
//  returned with the next non-comment line (whether or not it parses).
//
//  The globals InputRegExCount and InputExprErrors are set by this function.
func LoadFromScanner(efile *bufio.Scanner, f func(*RegExParsed)) []*RegExParsed {
	mpat := regexp.MustCompile(`^#(\w+): *(.*)`)
	elist := make([]*RegExParsed, 0)
	meta := make(map[string]string)
	InputRegExCount = 0
	InputErrorCount = 0
	for efile.Scan() {
		line := efile.Text()
		e := &RegExParsed{Expr: line}
		if IsComment(line) {
			r := mpat.FindStringSubmatch(line)
			if r != nil { // if recognized metadata format
				e.Meta = make(map[string]string)
				e.Meta[r[1]] = r[2]       // this line only
				addMeta(meta, r[1], r[2]) // also accumulate
			}
		} else {
			e.Tree, e.Err = Parse(line) // parse input
			if e.Tree != nil {          // if okay
				elist = append(elist, e) // save parse tree
				InputRegExCount++        // count success
			} else {
				InputErrorCount++ // else count error
			}
			e.Meta = meta                  // accumulated metadata
			meta = make(map[string]string) // reset meta collection
		}
		if f != nil {
			f(e)
		}
	}
	CkErr(efile.Err())
	return elist
}

// addMeta grows the metadata, concatenating with \n if the key is a duplicate.
func addMeta(meta map[string]string, key string, val string) {
	if meta[key] == "" {
		meta[key] = val
	} else {
		meta[key] = meta[key] + "\n" + val
	}
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
