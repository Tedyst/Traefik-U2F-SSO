package main

import (
	"embed"
	"net/http"
)

//go:embed static/*.html
var statics embed.FS

// Index is the main page where the user logs in or registers
func Index(w http.ResponseWriter, r *http.Request) {
	sess, err := sessionsstore.Get(r, "auth_session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// If the user is logged in, the page shown is static/loggedin.html
	if sess.Values["logged"] == true {
		logger.Debugw("User is not logged in",
			"Session", sess.ID,
		)
		newURL := r.URL.Query().Get("rd")
		if newURL != "" {
			http.Redirect(w, r, newURL, http.StatusSeeOther)
			return
		}

		render(w, "static/loggedin.html")
		return
	}
	if err := sess.Save(r, w); err != nil {
		logger.Errorf("error in peristing: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// If the registration is allowed in config.json, the page shown is static/index.html, that allows registration using a token.
	if Config.RegistrationAllowed {
		render(w, "static/index.html")
		return
	}
	// If this page is shown, only logging in using the authenticators is allowed
	render(w, "static/justlogin.html")
}

func render(w http.ResponseWriter, file string) {
	content, err := statics.ReadFile(file)
	if err != nil {
		logger.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	if _, err := w.Write(content); err != nil {
		logger.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
