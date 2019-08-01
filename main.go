package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/koesie10/webauthn/webauthn"
	"go.uber.org/zap"
)

//go:generate go run script/embedfiles.go

var (
	logger *zap.SugaredLogger
	// Config is imported from config.yml
	Config     Configuration
	webauth, _ = webauthn.New(&webauthn.Config{
		RelyingPartyName:   "webauthn-demo",
		AuthenticatorStore: storage,
		Debug:              true,
	})
)

func main() {
	initConfig()
	zaplog, _ := zap.NewProduction()
	if Config.Debug == true {
		zaplog, _ = zap.NewDevelopment()
	}
	defer zaplog.Sync()
	logger = zaplog.Sugar()

	// Prepare database
	var err error
	db, err = sql.Open("sqlite3", "storage/database?journal_mode=WAL")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	initStorage()

	logger.Info("Started on :", Config.Port)

	mux := http.NewServeMux()
	mux.HandleFunc("/", Index)
	mux.HandleFunc("/webauthn/registration/start", registrationStart)
	mux.HandleFunc("/webauthn/registration/finish", registrationFinish)
	mux.HandleFunc("/webauthn/login/start", loginStart)
	mux.HandleFunc("/webauthn/login/finish", loginFinish)
	mux.HandleFunc("/verify", verify)

	erra := http.ListenAndServe(":"+strconv.Itoa(Config.Port), RequestLogger(mux))
	if erra != nil {

		logger.Fatalf("Error in ListenAndServe: %s", erra)
	}
}

//RequestLogger logs every request
func RequestLogger(targetMux http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		targetMux.ServeHTTP(w, r)

		// log request by who(IP address)
		requesterIP := r.RemoteAddr

		logger.Infow("Loaded page",
			"Method", r.Method,
			"RequestURI", r.RequestURI,
			"RequesterIP", requesterIP,
			"Time", time.Since(start),
		)
	})
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
	sess.Save(r, w)
}

func registrationFinish(w http.ResponseWriter, r *http.Request) {
	if Config.RegistrationAllowed == false {
		http.Error(w, "Registration not allowed in config", http.StatusForbidden)
		logger.Debug("Registration attempt denied since the registration is disabled in config")
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

	logger.Debugw("Finishing registration",
		"Session", sess.ID,
		"User", u.Name,
	)
	webauth.FinishRegistration(r, w, u, webauthn.WrapMap(sess.Values))
}

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
