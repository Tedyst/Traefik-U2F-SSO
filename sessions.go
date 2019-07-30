
package main

import (
	"github.com/michaeljs1990/sqlitestore"
)

var (
	// key must be 16, 24 or 32 bytes long (AES-128, AES-192 or AES-256)
	key = []byte("super-secret-key")
	store, _ = sqlitestore.NewSqliteStore("./database", "sessions", "/", 3600, key)
)