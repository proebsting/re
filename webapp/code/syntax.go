//  syntax.go -- generate syntax explanation page

package webapp

import (
	"html/template"
	"net/http"
)

//  syntax generates a page outlining the accepted syntax
func syntax(w http.ResponseWriter, r *http.Request) {
	putheader(w, r, "Syntax")
	tSyntax.Execute(w, r)
	putfooter(w, r)
}

var tSyntax = template.Must(template.New("syntax").Parse(`
<P>We implement traditional regular expressions with a few simple extensions.
<P>The following forms are handled:<PRE>
      abc  a|b|c  a(b|c)d
      a?  b*  c+  d{m,n}
      \a \e \f \n \r \t \v \046 \xF7 \u03A8
      .  \d \s \w \D \S \W
      [abc]  [^abc]  [a-c]  [\x]
</PRE>
<P>We ignore the Perl non-capturing submatch form <CODE>(?:</CODE>,
but other <CODE>(?</CODE> forms are errors.
<P>All expressions are &ldquo;anchored&rdquo;.
An initial <CODE>^</CODE> and/or final <CODE>$</CODE> is ignored.
<P>Wildcard character sets (for
<CODE>&nbsp; . &nbsp; \w &nbsp; \D &nbsp; [^\d] &nbsp; </CODE> etc.)
are limited to the ASCII character set [\x01-\x7F].
`))
