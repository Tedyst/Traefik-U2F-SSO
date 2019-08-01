package main

import (
	"fmt"
	"net/http"
)

//Index is the main page
func Index(w http.ResponseWriter, r *http.Request) {
	sess, err := sessionsstore.Get(r, "auth_session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	sess.Save(r, w)
	// If the user is logged in, te page shown is static/loggedin.html
	if sess.Values["logged"] == true {
		fmt.Fprintf(w, loggedinHTML)
		return
	}
	// If the registration is allowed in config.json, te page shown is static/index.html, that allows registration using a token.
	if Config.RegistrationAllowed == true {
		fmt.Fprintf(w, indexHTML)
	} else {
		// If this page is shown, only logging in using the authenticators is allowed, the page being static/justlogin.html
		fmt.Fprintf(w, justloginHTML)
	}
}
