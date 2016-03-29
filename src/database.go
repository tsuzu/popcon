package main

import (
	"crypto/sha512"
	"database/sql"
	"encoding/hex"
	"errors"

	_ "github.com/go-sql-driver/mysql"
	"github.com/satori/go.uuid"
)

// Shared in all codes
var mainDB *DatabaseManager

// DatabaseManager is a connector to this database
type DatabaseManager struct {
	db *sql.DB
}

// UserAdd is a function to add a new user
// userID is the primary key
// userName is the unique key
// len(userID) <= 20, len(userName) <= 256 len(pass) <= 50, len(email) <= 50
func (dm *DatabaseManager) UserAdd(userID string, userName string, pass string, email string) error {
	if len(userID) > 20 {
		return errors.New("error: len(userID) > 20")
	}

	if len(userName) > 256 {
		return errors.New("error: len(userName) > 20")
	}

	if len(pass) > 50 {
		return errors.New("error: len(internalID) > 20")
	}
	passHashArr := sha512.Sum512([]byte(pass))
	//passHash := hex.EncodeToString(sha512.Sum512([]byte(pass)))

	if len(email) > 50 {
		return errors.New("error: len(internalID) > 20")
	}

	_, err := dm.db.Exec("insert into users (userID, userName, passHash, email, isMember) values (?, ?, ?, ?, false)", userID, userName, passHashArr[:], email)

	return err
}

// UserUpdate is a function to add a new user
// userID is the primary key
// userName is the unique key
// len(userID) <= 20, len(userName) <= 256 len(pass) <= 50, len(email) <= 50
func (dm *DatabaseManager) UserUpdate(userID string, userName string, pass string, email string) error {
	if len(userID) > 20 {
		return errors.New("error: len(userID) > 20")
	}

	if len(userName) > 256 {
		return errors.New("error: len(userName) > 20")
	}

	if len(pass) > 50 {
		return errors.New("error: len(internalID) > 20")
	}
	passHashArr := sha512.Sum512([]byte(pass))
	//	passHash := hex.EncodeToString(passHashArr[:])

	if len(email) > 50 {
		return errors.New("error: len(internalID) > 20")
	}

	_, err := dm.db.Exec("update users set userName=?, passHash=?, email=? where userID=?", userName, passHashArr[:], email, userID)

	return err
}

// User is a struct to save UserData
type User struct {
	userID   string
	userName string
	passHash [64]byte
	email    string
}

// UserFind is to return a User object
// userID is the primary key
// len(userID) <= 20
func (dm *DatabaseManager) UserFind(userID string) (*User, error) {
	if len(userID) > 20 {
		return nil, errors.New("error: len(internalID) > 20")
	}

	rows, err := dm.db.Query("select userName, passHash, email from users where userID=?", userID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var passHash []byte
	var userName, email string

	var user *User
	for rows.Next() {
		rows.Scan(&userName, &passHash, &email)

		user = &User{
			userID:   userID,
			userName: userName,
			email:    email,
		}

		for idx := range user.passHash {
			user.passHash[idx] = passHash[idx]
		}
	}

	if user == nil {
		return nil, errors.New("Not found")
	}

	return user, nil
}

// UserFindLight returns a User object which contains only userID and passHash
// userID is the primary key
// len(userID) <= 20
func (dm *DatabaseManager) UserFindLight(userID string) (*User, error) {
	if len(userID) > 20 {
		return nil, errors.New("error: len(internalID) > 20")
	}

	rows, err := dm.db.Query("select passHash from users where userID=?", userID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var passHash []byte

	var user *User
	for rows.Next() {
		rows.Scan(&passHash)

		user = &User{
			userID:   userID,
		}

		for idx := range user.passHash {
			user.passHash[idx] = passHash[idx]
		}
	}

	if user == nil {
		return nil, errors.New("Not found")
	}

	return user, nil
}

// SessionAdd is a function to add a new session
// len(userID) <= 20
func (dm *DatabaseManager) SessionAdd(userID string) (*string, error) {
	var err error

	cnt := 0
	for {
		uid := uuid.NewV4()
		sessionID := hex.EncodeToString(uid[:])

		_, err = dm.db.Exec("insert into sessions (sessionID, userID, unixTimeLimit) values(?, ?, unix_timestamp(now()) + 2592000)", sessionID, userID)

		if err == nil {
			return &sessionID, nil
		}

		if cnt > 3 {
			break
		}
		cnt++
	}

	return nil, err
}

// SessionFind is to find a session
// len(sessionID) = 32
func (dm *DatabaseManager) SessionFind(sessionID string) (*string, error) {
	rows, err := dm.db.Query("select userID from sessions where sessionID=?", sessionID)

	if err != nil {
		return nil, err
	}

	cnt := 0
	for rows.Next() {
        var uid string
		err = rows.Scan(&uid)
        
		if err == nil {
			return &uid, nil
		}

		if cnt > 3 {
			break
		}
		cnt++
	}
    
    return nil, errors.New("error: Not Found")
}

// SessionRemove is to remove session
// len(sessionID) = 32
func (dm *DatabaseManager) SessionRemove(sessionID string) error {
	_, err := dm.db.Exec("delete from sessions where sessionID=?", sessionID)

	return err
}

// NewDatabaseManager is a function to initialize database connections
// static function
func NewDatabaseManager() (*DatabaseManager, error) {
	dm := &DatabaseManager{}
	var err error

	// pcpjudge Database
	dm.db, err = sql.Open("mysql", "popcon:password@/popcon") // Should change password

	if err != nil {
		return nil, err
	}

	dm.db.SetMaxIdleConns(150)

	err = dm.db.Ping()

	if err != nil {
		return nil, err
	}

	// Create Users Table
	_, err = dm.db.Exec("create table if not exists users (userID varchar(20) primary key, userName varchar(256) unique, passHash varbinary(64), email varchar(50), isMember boolean)")

	if err != nil {
		return nil, err
	}

	// Create Sessions Table
	// TODO: Fix a bug about Year 2038 Bug in unixTimeLimit
	_, err = dm.db.Exec("create table if not exists sessions (sessionID varchar(50) primary key, userID varchar(20), unixTimeLimit int(11), index idx(unixTimeLimit))")

	if err != nil {
		return nil, err
	}

	return dm, nil
}
