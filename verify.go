package main

import "net/http"

func verify(w http.ResponseWriter, r *http.Request) {
	sess, err := sessionsstore.Get(r, "auth_session")
	if err != nil {
		http.Redirect(w, r, Config.URL, http.StatusSeeOther)
		logger.Debugw("Error getting the session",
			"Session", sess.ID,
		)
		return
	}

	if sess.Values["logged"] == true {
		logger.Debugw("User is logged in",
			"Session", sess.ID,
		)
		return

	}
	logger.Debugw("User is not logged in",
		"Session", sess.ID,
	)
	newURL := r.URL.Query().Get("rd")
	if newURL != "" {
		http.Redirect(w, r, Config.URL+"?rd="+newURL, http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, Config.URL, http.StatusSeeOther)

}
