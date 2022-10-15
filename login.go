package main

import (
	"encoding/json"
	"github.com/koesie10/webauthn/webauthn"
	"log"
	"net/http"
)

func loginStart(w http.ResponseWriter, r *http.Request) {
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

	logger.Debugw("Started logging in",
		"Session", sess.ID,
		"User", u.Name,
	)
	webauth.StartLogin(r, w, u, webauthn.WrapMap(sess.Values))
	sess.Save(r, w)
}

func loginFinish(w http.ResponseWriter, r *http.Request) {
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

	logger.Debugw("Finishing logging in",
		"Session", sess.ID,
		"User", u.Name,
	)
	authenticator := webauth.FinishLogin(r, w, u, webauthn.WrapMap(sess.Values))
	if authenticator == nil {
		logger.Debugw("Did not finish logging in",
			"Session", sess.ID,
			"User", u.Name,
		)
		return
	}

	_, ok := authenticator.(*Authenticator)
	if !ok {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logger.Debugw("Help",
			"Session", sess.ID,
			"User", u.Name,
		)
		return
	}

	logger.Debugw("Logged in",
		"Session", sess.ID,
		"User", u.Name,
	)

	payload, _ := json.Marshal(u)
	sess.Values["logged"] = true
	err = sess.Save(r, w)
	if err != nil {
		log.Fatal(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(payload)
}
