package main

import (
	"fmt"

	"database/sql"

	"github.com/Tedyst/sqlitestore"
	"github.com/koesie10/webauthn/webauthn"
	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

// User is needed for json reply when logged in.
type User struct {
	Name string `json:"name"`
}

// Authenticator is needed for webauthn protocol
type Authenticator struct {
	ID           []byte
	CredentialID []byte
	PublicKey    []byte
	AAGUID       []byte
	SignCount    uint32
}

// Storage is needed for webauthn protocol
type Storage struct {
}

var storage = &Storage{}

// AddAuthenticator is needed for webauthn protocol
func (s *Storage) AddAuthenticator(user webauthn.User, authenticator webauthn.Authenticator) error {
	stmt, err := db.Prepare("INSERT INTO authenticators(User, ID, CredentialID, PublicKey, AAGUID, SignCount) VALUES (?, ?, ?, ?, ?, ?)")
	if err != nil {
		logger.Error(err)
	}
	defer stmt.Close()
	_, err = stmt.Exec(user.WebAuthName(),
		authenticator.WebAuthID(),
		authenticator.WebAuthCredentialID(),
		authenticator.WebAuthPublicKey(),
		authenticator.WebAuthAAGUID(),
		authenticator.WebAuthSignCount())
	if err != nil {
		logger.Error(err)
		// return fmt.Errorf("authenticator already exists")
	}
	logger.Debugw("Added authenticator in database",
		"User", user.WebAuthName(),
		"AuthID", authenticator.WebAuthID(),
	)
	return nil
}

// GetAuthenticator is needed for webauthn protocol
func (s *Storage) GetAuthenticator(id []byte) (webauthn.Authenticator, error) {
	var au Authenticator
	var user string
	stmt, err := db.Prepare("SELECT User,ID,CredentialID,PublicKey,AAGUID,SignCount FROM authenticators WHERE ID = ?")
	if err != nil {
		logger.Error(err)
	}
	rows, err := stmt.Query(id)
	defer rows.Close()
	defer stmt.Close()
	for rows.Next() {
		err = rows.Scan(&user, &au.ID, &au.CredentialID, &au.PublicKey, &au.AAGUID, &au.SignCount)
		if err != nil {
			logger.Error(err)
		}
		logger.Debugw("Found authenticator in database",
			"User", user,
			"AuthID", id,
		)
		return &au, nil
	}
	err = rows.Err()
	if err != nil {
		logger.Error(err)
	}
	logger.Debugw("Did not find authenticator in database",
		"AuthID", id,
	)
	return nil, fmt.Errorf("authenticator not found")
}

// GetAuthenticators is needed for webauthn protocol
func (s *Storage) GetAuthenticators(user webauthn.User) ([]webauthn.Authenticator, error) {
	var authrs []webauthn.Authenticator
	stmt, err := db.Prepare("SELECT ID, CredentialID, PublicKey, AAGUID, SignCount FROM authenticators WHERE User = ?")
	if err != nil {
		logger.Error(err)
	}
	rows, err := stmt.Query(user.WebAuthName())
	if err != nil {
		logger.Error(err)
		return authrs, nil
	}
	defer rows.Close()
	defer stmt.Close()
	for rows.Next() {
		var au Authenticator
		err = rows.Scan(&au.ID, &au.CredentialID, &au.PublicKey, &au.AAGUID, &au.SignCount)
		if err != nil {
			logger.Error(err)
		}
		logger.Debugw("Found authenticator in database",
			"User", user.WebAuthName(),
			"AuthID", au.ID,
		)
		authrs = append(authrs, &au)
	}
	err = rows.Err()
	if err != nil {
		logger.Error(err)
	}
	return authrs, nil
}

// WebAuthID is needed for webauthn protocol
func (u *User) WebAuthID() []byte {
	return []byte(u.Name)
}

// WebAuthName is needed for webauthn protocol
func (u *User) WebAuthName() string {
	return u.Name
}

// WebAuthDisplayName is needed for webauthn protocol
func (u *User) WebAuthDisplayName() string {
	return u.Name
}

// WebAuthID is needed for webauthn protocol
func (a *Authenticator) WebAuthID() []byte {
	return a.ID
}

// WebAuthCredentialID is needed for webauthn protocol
func (a *Authenticator) WebAuthCredentialID() []byte {
	return a.CredentialID
}

// WebAuthPublicKey is needed for webauthn protocol
func (a *Authenticator) WebAuthPublicKey() []byte {
	return a.PublicKey
}

// WebAuthAAGUID is needed for webauthn protocol
func (a *Authenticator) WebAuthAAGUID() []byte {
	return a.AAGUID
}

// WebAuthSignCount is needed for webauthn protocol
func (a *Authenticator) WebAuthSignCount() uint32 {
	return a.SignCount
}

func initStorage() {
	var err error
	// Test storage
	err = db.Ping()
	if err != nil {
		logger.Error(err)
	}

	sessionsstore, _ = sqlitestore.NewSqliteStoreFromConnection(db, "sessions", "/", 360000, sessionskey)
	sessionsstore.Options.Domain = Config.Domain
	sessionsstore.Options.Secure = true
	sessionsstore.Options.HttpOnly = false

	db.Exec("PRAGMA journal_mode=WAL")

	sqlStmt := `CREATE TABLE IF NOT EXISTS users (id integer not null primary key, name text);
	CREATE TABLE IF NOT EXISTS authenticators (
		User TEXT,
		ID BLOB UNIQUE,
		CredentialID BLOB,
		PublicKey BLOB,
		AAGUID BLOB,
		SignCount INTEGER
	);`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		logger.Error(err)
		return
	}
}
