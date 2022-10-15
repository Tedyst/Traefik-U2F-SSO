package main

import (
	"github.com/koesie10/webauthn/webauthn"
	"net/http"
)

func registrationFinish(w http.ResponseWriter, r *http.Request) {
	if Config.RegistrationAllowed == false {
		http.Error(w, "Registration not allowed in config", http.StatusForbidden)
		logger.Debug("Registration attempt denied since the registration is disabled in config")
		return
	}
	u := &User{
		Name: r.URL.Query().Get("name"),
	}

	// TODO sja: if domain mismatches, we have an error here, log that!
	sess, err := sessionsstore.Get(r, "auth_session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logger.Errorw("Error getting a session",
			"Session", sess.ID,
			"User", u.Name,
		)
		return
	}

	logger.Debugw("Finishing registration",
		"Session", sess.ID,
		"User", u.Name,
	)
	webauth.FinishRegistration(r, w, u, webauthn.WrapMap(sess.Values))
}

func registrationStart(w http.ResponseWriter, r *http.Request) {
	if Config.RegistrationAllowed == false {
		http.Error(w, "Registration not allowed in config", http.StatusForbidden)
		logger.Debug("Registration attempt denied since the registration is disabled in config")
		return
	}
	if Config.RegistrationToken != r.URL.Query().Get("token") {
		http.Error(w, "Wrong token", http.StatusForbidden)
		logger.Debug("Registration attempt denied since the token is wrong")
		return
	}
	u := &User{
		Name: r.URL.Query().Get("name"),
	}

	sess, err := sessionsstore.Get(r, "auth_session")

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logger.Errorw("Error getting a session",
			"Session", sess.ID,
			"User", u.Name,
		)
		return
	}

	logger.Debugw("Started registration",
		"Session", sess.ID,
		"User", u.Name,
	)
	webauth.StartRegistration(r, w, u, webauthn.WrapMap(sess.Values))

	if err := sess.Save(r, w); err != nil {
		logger.Error("error persiting registration: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
