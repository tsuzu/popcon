package main

import "errors"
import "crypto/sha512"
import "strconv"

// User is a struct to save UserData
// "create table if not exists users (internalID int(11) auto_increment primary key, userID varchar(20) unique, userName varchar(256) unique, passHash varbinary(64), email varchar(50), groupID int(11))"
type User struct {
	Iid int64 `db:"pk"`
	Uid     string `default:""`
	UserName   string `default:""`
	PassHash   []byte 
	Email      string `default:""`
	Gid  int64 `default:""`
}

func (dm *DatabaseManager) CreateUserTable() error {
	err := dm.db.CreateTableIfNotExists(&User{})

	if err != nil {
		return err
	}

	dm.db.CreateUniqueIndex(&User{}, "uid")
	dm.db.CreateUniqueIndex(&User{}, "user_name")
	dm.db.CreateUniqueIndex(&User{}, "email")
	
	return err
}

// UserAdd is a function to add a new user
// userID is the primary key
// userName is the unique key
// len(userID) <= 20, len(userName) <= 256 len(pass) <= 50, len(email) <= 50
func (dm *DatabaseManager) UserAdd(userID string, userName string, pass string, email string, groupID int64) (int64, error) {
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

	return dm.db.Insert(&User{
		Uid: userID,
		UserName: userName,
		PassHash: passHashArr[:],
		Email: email,
		Gid: groupID,
	})
}

// UserUpdate is a function to add a new user
// len(userID) <= 20, len(userName) <= 256 len(pass) <= 50, len(email) <= 50
func (dm *DatabaseManager) UserUpdate(internalID int, userID string, userName string, pass string, email string, groupID int64) error {
	if len(userID) > 20 {
		return errors.New("error: len(userID) > 20")
	}

	if len(userName) > 256 {
		return errors.New("error: len(userName) > 256")
	}

	if len(pass) > 50 {
		return errors.New("error: len(pass) > 50")
	}
	passHashArr := sha512.Sum512([]byte(pass))

	if len(email) > 50 {
		return errors.New("error: len(email) > 50")
	}

	_, err := dm.db.Update(&User{
		Uid: userID,
		UserName: userName,
		PassHash: passHashArr[:],
		Email: email,
		Gid: groupID,
	})

	return err
}

// UserFind is to return a User object
// userID is the unique key
// len(userID) <= 20
func (dm *DatabaseManager) userFind(key string, value string) (*User, error) {
	var resulsts []User
	err := dm.db.Select(&resulsts, dm.db.Where(key, "=", value))
	
	if err != nil {
		return nil, err
	}

	if len(resulsts) == 0 {
		return nil, errors.New("Unknown user")
	}

	return &resulsts[0], nil
}

// UserFindFromIID is to return a User object
func (dm *DatabaseManager) UserFindFromIID(internalID int64) (*User, error) {
	return dm.userFind("iid", strconv.FormatInt(internalID, 10))
}

// UserFindFromUserID is to return a User object
func (dm *DatabaseManager) UserFindFromUserID(userID string) (*User, error) {
	if len(userID) > 20 {
		return nil, errors.New("error: len(userID) > 20")
	}

	return dm.userFind("uid", userID)

}
