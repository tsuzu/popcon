package main

import "errors"
import "encoding/hex"
import "github.com/satori/go.uuid"

// SessionTemplateData contains data in SQL DB
type SessionTemplateData struct {
	IsSignedIn bool
	UserID     string
	UserName   string
}

// GetSessionTemplateData returns a SessionTemplateData object
func GetSessionTemplateData(sessionID string) (*SessionTemplateData, error) {
	user, err := GetSessionUserData(sessionID)

	if err != nil {
		return nil, err
	}

	return &SessionTemplateData{true, user.UserID, user.UserName}, nil
}

// GetSessionUserData returns an User object
func GetSessionUserData(sessionID string) (*User, error) {
	internalID, err := mainDB.SessionFind(sessionID)

	if err != nil {
		return nil, err
	}

	return mainDB.UserFindFromInternalID(internalID)
}

// SessionAdd adds a new session
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

//TODO implement caches of sessions
/*
type SessionTmplDataChan chan *SessionTemplateData
func getSession(c chan struct {SessionTmplDataChan; string}) {

}

func InitSessionManager() {
    go func() {

    }
}

*/
