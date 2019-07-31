package main

import (
	"github.com/koesie10/webauthn/webauthn"
	"log"
	"net/http"
	"time"
)

//go:generate go run script/embedfiles.go

var webauth, _ = webauthn.New(&webauthn.Config{
    RelyingPartyName:   "webauthn-demo",
	AuthenticatorStore: storage,
	Debug: true,
})

func main() {
	initStorage()
	log.Print("Started on :8080")
	mux := http.NewServeMux()
	mux.HandleFunc("/", Index)
	mux.HandleFunc("/webauthn/registration/start", registrationStart)
	mux.HandleFunc("/webauthn/registration/finish", registrationFinish)

	err := http.ListenAndServe(":8080", RequestLogger(mux))
	if err != nil {
		log.Fatalf("Error in ListenAndServe: %s", err)
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

func registrationStart (w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	u, ok := storage.users[name]
	if !ok {
		u = &User{
			Name:           name,
			Authenticators: make(map[string]*Authenticator),
		}
		storage.users[name] = u
	}

	sess, err := store.Get(r, "session")
	
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	webauth.StartRegistration(r, w, u, webauthn.WrapMap(sess.Values))
	sess.Save(r, w)
}

func registrationFinish (w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	u, ok := storage.users[name]
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
    	w.Write([]byte("500 - Something bad happened!"))
	}

	sess, err := store.Get(r, "session")
	if err != nil{
		sess.Save(r,w)
	}

	webauth.FinishRegistration(r, w, u, webauthn.WrapMap(sess.Values))
}