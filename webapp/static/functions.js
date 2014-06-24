<!--

//  atcs() -- Generate HTML link for mailing to "uname@cs.arizona.edu".
//  Use the label if supplied, else label with the e-mail address.

function atcs(uname, label, subj) {
	var addr = uname + '@cs.arizona.edu';
	if (!label)
		label = addr;
	if (subj)
		subj = '?subject=' + subj;
	else
		subj = '';
	document.write('<a href="mailto:' + addr + subj + '">' + label +'</a>');
}

//  clearForm(f) -- clear (not just reset) the contents of a form
//
//  based on code downloaded 16-May-2014
//  from www.javascript-coder.com/javascript-form/javascript-reset-form.phtml

function clearForm(oForm) {
	var elements = oForm.elements; 
	oForm.reset();
	for (i=0; i<elements.length; i++) {
		field_type = elements[i].type.toLowerCase();
		switch(field_type) {
			case "text": 
			case "password": 
			case "textarea":
			case "hidden":	
				elements[i].value = ""; 
				break;
			case "radio":
			case "checkbox":
				if (elements[i].checked) {
					elements[i].checked = false; 
				}
				break;
			case "select-one":
			case "select-multi":
				elements[i].selectedIndex = -1;
				break;
			default: 
				break;
		}
	}
}

//-->
