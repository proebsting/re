//  syntax.go -- generate syntax explanation page

package webapp

import (
	"html/template"
	"net/http"
)

//  syntax generates a page outlining the accepted syntax
func syntax(w http.ResponseWriter, r *http.Request) {
	putheader(w, r, "Syntax")
	tSyntax.Execute(w, nil)
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
<P>All expressions are &ldquo;anchored&rdquo;.
An initial <CODE>^</CODE> and/or final <CODE>$</CODE> is ignored.
<BR>Embedded anchors, as well as other anchor forms like <CODE>\b</CODE>,
are illegal.
<P>Wildcard character sets (for
<CODE>&nbsp; . &nbsp; \w &nbsp; \D &nbsp; [^\d] &nbsp; </CODE> etc.)
are limited to the ASCII subset [\x01-\x7F] of Unicode.
<P>Most Perl <CODE>(?</CODE> forms are illegal,
but two kinds are allowed and ignored:<PRE>
      Comments:  (?#...)
      Captures:  (?:  (?'id'  (?&lt;id&gt;  (?P&lt;id&gt;
</PRE>
<P>Leading and trailing spaces are trimmed to tolerate imprecision in
cutting and pasting.  To inhibit trimming, parenthesize the expression.
`))
