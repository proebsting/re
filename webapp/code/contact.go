//  contact.go -- generate "contact" page

//  'contact' is not a verb in this house.
//	-- Nero Wolfe

package webapp

import (
	"fmt"
	"net/http"
)

//  contact generates a page with a mailto: link for feedback
func contact(w http.ResponseWriter, r *http.Request) {
	putheader(w, r, "Contact Us")
	fmt.Fprintf(w, `
<P> We love feedback.
<P> Send mail to %s.
`, MAILTO)
	putfooter(w, r)
}
