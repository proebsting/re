<!--

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

// pretect(s) -- wrap in <pre>...</pre> and escape [&<>"] characters

function pretect(s) {
	return "<pre>" + s.replace(/\&/g, "&amp;").replace(/\"/g, "&quot;").
		replace(/</g, "&lt;").replace(/>/g, "&gt;") + "</pre>";
}

//-->
