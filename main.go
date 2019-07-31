package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/koesie10/webauthn/webauthn"
)

//go:generate go run script/embedfiles.go

var webauth, _ = webauthn.New(&webauthn.Config{
	RelyingPartyName:   "webauthn-demo",
	AuthenticatorStore: storage,
	Debug:              true,
})

func main() {
	// Prepare database
	var err error
	db, err = sql.Open("sqlite3", "database?journal_mode=WAL")
	if err != nil {
		log.Fatal(err)
	}
	initStorage()
	log.Print("Started on :8080")
	mux := http.NewServeMux()
	mux.HandleFunc("/", Index)
	mux.HandleFunc("/webauthn/registration/start", registrationStart)
	mux.HandleFunc("/webauthn/registration/finish", registrationFinish)
	mux.HandleFunc("/webauthn/login/start", loginStart)
	mux.HandleFunc("/webauthn/login/finish", loginFinish)
	mux.HandleFunc("/verify", loggedIn)

	erra := http.ListenAndServe(":8080", RequestLogger(mux))
	if erra != nil {
		log.Fatalf("Error in ListenAndServe: %s", erra)
	}
}

//RequestLogger logs every request
func RequestLogger(targetMux http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		targetMux.ServeHTTP(w, r)

		// log request by who(IP address)
		requesterIP := r.RemoteAddr

		log.Printf(
			"%s\t\t%s\t\t%s\t\t%v",
			r.Method,
			r.RequestURI,
			requesterIP,
			time.Since(start),
		)
	})
}

func registrationStart(w http.ResponseWriter, r *http.Request) {
	u := &User{
		Name: r.URL.Query().Get("name"),
	}

	sess, err := sessionsstore.Get(r, "session")

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	webauth.StartRegistration(r, w, u, webauthn.WrapMap(sess.Values))
	sess.Save(r, w)
}

func registrationFinish(w http.ResponseWriter, r *http.Request) {
	u := &User{
		Name: r.URL.Query().Get("name"),
	}

	sess, err := sessionsstore.Get(r, "session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	webauth.FinishRegistration(r, w, u, webauthn.WrapMap(sess.Values))
}

func loginStart(w http.ResponseWriter, r *http.Request) {
	u := &User{
		Name: r.URL.Query().Get("name"),
	}

	sess, err := sessionsstore.Get(r, "session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	webauth.StartLogin(r, w, u, webauthn.WrapMap(sess.Values))
	sess.Save(r, w)
}

func loginFinish(w http.ResponseWriter, r *http.Request) {
	u := &User{
		Name: r.URL.Query().Get("name"),
	}

	sess, err := sessionsstore.Get(r, "session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	authenticator := webauth.FinishLogin(r, w, u, webauthn.WrapMap(sess.Values))
	if authenticator == nil {
		return
	}

	_, ok := authenticator.(*Authenticator)
	if !ok {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Print(u.Name + " logged in!")
	payload, _ := json.Marshal(u)
	sess.Values["logged"] = true
	err = sess.Save(r, w)
	if err != nil {
		log.Fatal(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(payload)
}

func loggedIn(w http.ResponseWriter, r *http.Request) {
	sess, err := sessionsstore.Get(r, "session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if sess.Values["logged"] == true {
		fmt.Fprintf(w, "logged in.")
	} else {
		fmt.Fprint(w, "not logged in.")
	}
}
