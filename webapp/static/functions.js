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
	for (var i = 0; i < elements.length; i++) {
		var field_type = elements[i].type.toLowerCase();
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

//  addslot(tx) -- add one input field to sequence beginning with tabindex tx
//
//  very sensitive to the HTML structure generated by putform() in the app code

function addslot(tx, maxn) {	
	var elast, e;
	maxn += tx
	while ((e = document.querySelector("[tabindex='" +tx + "']")) != null) {
		elast = e;
		tx = tx + 1;
	}
	if (tx >= maxn)
		return;
	var dlast = elast.parentElement;	// enclosing <div>
	var dnew = dlast.cloneNode(false);	// copy <div>
	var enew = elast.cloneNode(true);	// copy <input>
	enew.setAttribute("name", "v" + tx);	// name = "v" + tabindex
	enew.setAttribute("tabindex", tx);
	enew.value = "";
	dnew.appendChild(enew);
	dlast.parentElement.insertBefore(dnew, dlast.nextSibling);
}

//-->
