package main

import "errors"
import "crypto/sha512"

// User is a struct to save UserData
type User struct {
	InternalID int
	UserID     string
	UserName   string
	PassHash   [64]byte
	Email      string
	GroupID  int
}

// UserAdd is a function to add a new user
// userID is the primary key
// userName is the unique key
// len(userID) <= 20, len(userName) <= 256 len(pass) <= 50, len(email) <= 50
func (dm *DatabaseManager) UserAdd(userID string, userName string, pass string, email string, groupID int) (int, error) {
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

	res, err := dm.db.Exec("insert into users (userID, userName, passHash, email, groupID) values (?, ?, ?, ?, ?)", userID, userName, passHashArr[:], email, groupID)

	if err != nil {
		return 0, err
	} else if last, err := res.LastInsertId(); err == nil {
		return int(last), err
	} else {
		return 0, err
	}
}

// UserUpdate is a function to add a new user
// len(userID) <= 20, len(userName) <= 256 len(pass) <= 50, len(email) <= 50
func (dm *DatabaseManager) UserUpdate(internalID int, userID string, userName string, pass string, email string, groupID int) error {
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

	_, err := dm.db.Exec("update users set userID=? userName=?, passHash=?, email=?, groupID=? where internalID=?", userID, userName, passHashArr[:], email, groupID, internalID)

	return err
}

// UserFind is to return a User object
// userID is the unique key
// len(userID) <= 20
func (dm *DatabaseManager) UserFind(userID string) (*User, error) {
	if len(userID) > 20 {
		return nil, errors.New("error: len(userID) > 20")
	}

	rows, err := dm.db.Query("select internalID, userName, passHash, email, &groupID from users where userID=?", userID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var passHash []byte
	var userName, email string
	var internalID, groupID int

	var user *User
	for rows.Next() {
		rows.Scan(&internalID, &userName, &passHash, &email, &groupID)

		user = &User{
			UserID:     userID,
			UserName:   userName,
			Email:      email,
			InternalID: internalID,
            GroupID: groupID,
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
	rows, err := dm.db.Query("select userID, userName, passHash, email, groupID from users where internalID=?", internalID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var passHash []byte
	var userID, userName, email string
	var groupID int

	var user *User
	for rows.Next() {
		rows.Scan(&userID, &userName, &passHash, &email, &groupID)

		user = &User{
			InternalID: internalID,
			UserID:     userID,
			UserName:   userName,
			Email:      email,
			GroupID: groupID,
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

// Group is a struct to save GroupData
type Group struct {
    GroupID int
    GroupName string
}

// GroupAdd adds a new group
// len(groupName) <= 50
func (dm *DatabaseManager) GroupAdd(groupName string) (int, error) {
    if len(groupName) > 50 {
        return 0, errors.New("len(groupName) > 50")
    }
    
    res, err := dm.db.Exec("insert into groups (groupName) values (?)", groupName)
    
    if err != nil {
        return 0, err
    }
    
    if id, err := res.LastInsertId(); err != nil {
        return 0, err
    } else {
        return int(id), err
    }
}

// GroupFind finds a group with groupID
func (dm *DatabaseManager) GroupFind(groupID int) (*string, error) {
    
    res, err := dm.db.Query("select groupName from groups where groupID=?", groupID)
    
    if err != nil {
        return nil, err
    }
    
    var groupName string

    if !res.Next() {
        return nil, errors.New("rows.Next() failed")
    }
    
    err = res.Scan(&groupName)
    
    if err != nil {
        return nil, err
    }
    
    return &groupName, nil
}

// GroupRemove removes from groups
func (dm *DatabaseManager) GroupRemove(groupID int) error {
    _, err := dm.db.Exec("delete from groups where groupID=?", groupID)

	return err 
}