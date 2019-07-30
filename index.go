package main

import (
	"fmt"
	"net/http"
)

//Index is the main page
func Index(w http.ResponseWriter, r *http.Request) {
	sess, err := store.Get(r, "cookie-name")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	sess.Save(r, w)

	fmt.Fprintf(w, indexHTML)
}
