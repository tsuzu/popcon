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
func (dm *DatabaseManager) UserAdd(userID string, userName string, pass string, email string) (int, error) {
	if len(userID) > 20 {
		return 0, errors.New("error: len(userID) > 20")
	}

	if len(userName) > 256 {
		return 0, errors.New("error: len(userName) > 256")
	}

	if len(pass) > 50 {
		return 0, errors.New("error: len(pass) > 50")
	}
	passHashArr := sha512.Sum512([]byte(pass))
	//passHash := hex.EncodeToString(sha512.Sum512([]byte(pass)))

	if len(email) > 50 {
		return 0, errors.New("error: len(email) > 50")
	}

	res, err := dm.db.Exec("insert into users (userID, userName, passHash, email, isMember) values (?, ?, ?, ?, false)", userID, userName, passHashArr[:], email)

	if err != nil {
		return 0, err
	} else if last, err := res.LastInsertId(); err == nil {
		return int(last), err
	} else {
		return 0, err
	}
}

// UserUpdate is a function to add a new user
// userID is the primary key
// userName is the unique key
// len(userID) <= 20, len(userName) <= 256 len(pass) <= 50, len(email) <= 50
func (dm *DatabaseManager) UserUpdate(internalID, userID string, userName string, pass string, email string) error {
	if len(userID) > 20 {
		return errors.New("error: len(userID) > 20")
	}

	if len(userName) > 256 {
		return errors.New("error: len(userName) > 256")
	}

	if len(pass) > 50 {
		return errors.New("error: len(internalID) > 50")
	}
	passHashArr := sha512.Sum512([]byte(pass))
	//	passHash := hex.EncodeToString(passHashArr[:])

	if len(email) > 50 {
		return errors.New("error: len(internalID) > 50")
	}

	_, err := dm.db.Exec("update users set userName=?, passHash=?, email=? where userID=?", userName, passHashArr[:], email, userID)

	return err
}

// UserFind is to return a User object
// userID is the unique key
// len(userID) <= 20
func (dm *DatabaseManager) UserFind(userID string) (*User, error) {
	if len(userID) > 20 {
		return nil, errors.New("error: len(userID) > 20")
	}

	rows, err := dm.db.Query("select internalID, userName, passHash, email from users where userID=?", userID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var passHash []byte
	var userName, email string
	var internalID int

	var user *User
	for rows.Next() {
		rows.Scan(&internalID, &userName, &passHash, &email)

		user = &User{
			UserID:     userID,
			UserName:   userName,
			Email:      email,
			InternalID: internalID,
		}

		for idx := range user.PassHash {
			user.PassHash[idx] = passHash[idx]
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

	rows, err := dm.db.Query("select internalID, passHash from users where userID=?", userID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var passHash []byte
	var internalID int

	var user *User
	for rows.Next() {
		rows.Scan(&internalID, &passHash)

		user = &User{
			UserID:     userID,
			InternalID: internalID,
		}

		for idx := range user.PassHash {
			user.PassHash[idx] = passHash[idx]
		}
	}

	if user == nil {
		return nil, errors.New("Not found")
	}

	return user, nil
}

// UserFindFromInternalID is to return a User object
func (dm *DatabaseManager) UserFindFromInternalID(internalID int) (*User, error) {
	rows, err := dm.db.Query("select userID, userName, passHash, email from users where internalID=?", internalID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var passHash []byte
	var userID, userName, email string

	var user *User
	for rows.Next() {
		rows.Scan(&userID, &userName, &passHash, &email)

		user = &User{
			InternalID: internalID,
			UserID:     userID,
			UserName:   userName,
			Email:      email,
		}

		for idx := range user.PassHash {
			user.PassHash[idx] = passHash[idx]
		}
	}

	if user == nil {
		return nil, errors.New("Not found")
	}

	return user, nil
}

// SessionAdd is a function to add a new session
func (dm *DatabaseManager) SessionAdd(internalID int) (*string, error) {
	var err error

	cnt := 0
	for {
		uid := uuid.NewV4()
		sessionID := hex.EncodeToString(uid[:])

		_, err = dm.db.Exec("insert into sessions (sessionID, internalID, unixTimeLimit) values(?, ?, unix_timestamp(now()) + 2592000)", sessionID, internalID)

		if err == nil {
			return &sessionID, nil
		}

		if cnt > 3 {
			break
		}
		cnt++
	}

	return nil, errors.New("Failed to insert a new session(" + err.Error() + ")")
}

// SessionFind is to find a session
// len(sessionID) = 32
func (dm *DatabaseManager) SessionFind(sessionID string) (int, error) {
	rows, err := dm.db.Query("select internalID from sessions where sessionID=?", sessionID)

	if err != nil {
		return 0, err
	}

	cnt := 0
	for rows.Next() {
		var internalID int
		err = rows.Scan(&internalID)

		if err == nil {
			return internalID, nil
		}

		if cnt > 3 {
			break
		}
		cnt++
	}

	return 0, errors.New("error: Not Found")
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
	_, err = dm.db.Exec("create table if not exists users (internalID int(11) auto_increment primary key, userID varchar(20) unique, userName varchar(256) unique, passHash varbinary(64), email varchar(50), isMember boolean, authortity int(11))")

	if err != nil {
		return nil, err
	}

	// Create Sessions Table
	// TODO: Fix a bug about Year 2038 Bug in unixTimeLimit
	_, err = dm.db.Exec("create table if not exists sessions (sessionID varchar(50) primary key, internalID int(11), unixTimeLimit int(11), index iid(internalID), index idx(unixTimeLimit))")

	if err != nil {
		return nil, err
	}

	return dm, nil
}
