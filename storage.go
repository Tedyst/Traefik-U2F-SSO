package main

import (
	"fmt"
	"log"

	"database/sql"

	"github.com/koesie10/webauthn/webauthn"
	_ "github.com/mattn/go-sqlite3"
)

var (
	db *sql.DB
)

//User is the user type
type User struct {
	Name           string                    `json:"name"`
	Authenticators map[string]*Authenticator `json:"-"`
}

type Authenticator struct {
	ID           []byte
	CredentialID []byte
	PublicKey    []byte
	AAGUID       []byte
	SignCount    uint32
}

type Storage struct {
	users          map[string]*User
	authenticators map[string]*Authenticator
}

var storage = &Storage{
	authenticators: make(map[string]*Authenticator),
	users:          make(map[string]*User),
}

func (s *Storage) AddAuthenticator(user webauthn.User, authenticator webauthn.Authenticator) error {
	stmt, err := db.Prepare("INSERT INTO authenticators(User, ID, CredentialID, PublicKey, AAGUID, SignCount) VALUES (?, ?, ?, ?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	_, err = stmt.Exec(user.WebAuthName,
		authenticator.WebAuthID(),
		authenticator.WebAuthCredentialID(),
		authenticator.WebAuthPublicKey(),
		authenticator.WebAuthAAGUID(),
		authenticator.WebAuthSignCount())
	if err != nil {
		log.Fatal(err)
		// return fmt.Errorf("authenticator already exists")
	}

	return nil
}

func (s *Storage) GetAuthenticator(id []byte) (webauthn.Authenticator, error) {
	var au Authenticator
	stmt, err := db.Prepare("SELECT ID,CredidentialID,PublicKey,AAGUID,SignCount FROM authenticators WHERE ID = ?")
	if err != nil {
		log.Fatal(err)
	}
	rows, err := stmt.Query(id)
	defer stmt.Close()
	for rows.Next() {
		err = rows.Scan(au.ID, au.CredentialID, au.PublicKey, au.AAGUID, au.SignCount)
		if err != nil {
			log.Fatal(err)
		}
		return &au, nil
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	return nil, fmt.Errorf("authenticator not found")
}

func (s *Storage) GetAuthenticators(user webauthn.User) ([]webauthn.Authenticator, error) {
	var authrs []webauthn.Authenticator
	stmt, err := db.Prepare("SELECT ID,CredidentialID,PublicKey,AAGUID,SignCount FROM authenticators WHERE User = ?")
	if err != nil {
		log.Fatal(err)
	}
	rows, err := stmt.Query(user.WebAuthName)
	defer stmt.Close()
	for rows.Next() {
		var au Authenticator
		err = rows.Scan(au.ID, au.CredentialID, au.PublicKey, au.AAGUID, au.SignCount)
		if err != nil {
			log.Fatal(err)
		}
		authrs = append(authrs, &au)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	if len(authrs) == 0 {
		return nil, fmt.Errorf("authenticator not found")
	}
	return authrs, nil
}

func (u *User) WebAuthID() []byte {
	return []byte(u.Name)
}

func (u *User) WebAuthName() string {
	return u.Name
}

func (u *User) WebAuthDisplayName() string {
	return u.Name
}

func (a *Authenticator) WebAuthID() []byte {
	return a.ID
}

func (a *Authenticator) WebAuthCredentialID() []byte {
	return a.CredentialID
}

func (a *Authenticator) WebAuthPublicKey() []byte {
	return a.PublicKey
}

func (a *Authenticator) WebAuthAAGUID() []byte {
	return a.AAGUID
}

func (a *Authenticator) WebAuthSignCount() uint32 {
	return a.SignCount
}

func initStorage() {
	// Prepare database
	db, err := sql.Open("sqlite3", "database")
	if err != nil {
		log.Fatal(err)
	}

	sqlStmt := `
	CREATE TABLE IF NOT EXISTS users (id integer not null primary key, name text);
	CREATE TABLE IF NOT EXISTS authenticators (
		User TEXT,
		ID BLOB UNIQUE,
		CredidentialID BLOB,
		PublicKey BLOB,
		AAGUID BLOB,
		SignCount INTEGER
	);
	`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return
	}
}
