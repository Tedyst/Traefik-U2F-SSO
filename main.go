package main

import (
	"github.com/koesie10/webauthn/webauthn"
	"log"
	"net/http"
	"strconv"
)

var webauth, _ = webauthn.New(&webauthn.Config{
	RelyingPartyName:   "webauthn-demo",
	AuthenticatorStore: storage,
	Debug:              true,
})

func main() {
	if err := initConfig(); err != nil {
		log.Fatalf("could not init config. %w", err)
	}
	if err := initLogger(Config); err != nil {
		log.Fatalf("could not init logger. %w", err)
	}
	if err := initStorage(Config); err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	logger.Info("Started on :", Config.Port)

	mux := http.NewServeMux()
	mux.HandleFunc("/", Index)
	mux.HandleFunc("/webauthn/registration/start", registrationStart)
	mux.HandleFunc("/webauthn/registration/finish", registrationFinish)
	mux.HandleFunc("/webauthn/login/start", loginStart)
	mux.HandleFunc("/webauthn/login/finish", loginFinish)
	mux.HandleFunc("/verify", verify)

	if err := http.ListenAndServe(":"+strconv.Itoa(Config.Port), RequestLogger(mux)); err != nil {
		logger.Fatalf("Error in ListenAndServe: %s", err)
	}
}
