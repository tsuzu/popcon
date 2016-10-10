package main

import "errors"
import "encoding/hex"
import "github.com/satori/go.uuid"
import "time"

// SessionTemplateData contains data in SQL DB
//"create table if not exists sessions (sessionKey varchar(50) primary key, internalID int(11), unixTimeLimit int(11), index iid(internalID), index idx(unixTimeLimit))"
type Session struct {
	SessionKey string `db:"pk" default:"" size:"50"`
	Iid        int64  `default:""`
	TimeLimit  int64  `default:""`
}

type SessionTemplateData struct {
	IsSignedIn bool
	Iid        int64
	UserID     string
	UserName   string
	Gid        int64
}

func (dm *DatabaseManager) CreateSessionTable() error {
	err := dm.db.CreateTableIfNotExists(&Session{})

	if err != nil {
		return err
	}

	dm.db.CreateIndex(&Session{}, "iid")
	dm.db.CreateIndex(&Session{}, "timeLimit")

	return nil
}

// GetSessionTemplateData returns a SessionTemplateData object
func GetSessionTemplateData(sessionKey string) (*SessionTemplateData, error) {
	user, err := GetSessionUserData(sessionKey)

	if err != nil {
		return nil, err
	}

	return &SessionTemplateData{true, user.Iid, user.Uid, user.UserName, user.Gid}, nil
}

// GetSessionUserData returns an User object
func GetSessionUserData(sessionID string) (*User, error) {
	session, err := mainDB.SessionFind(sessionID)

	if err != nil {
		return nil, err
	}

	return mainDB.UserFindFromIID(session.Iid)
}

// SessionAdd adds a new session
func (dm *DatabaseManager) SessionAdd(internalID int64) (*string, error) {
	var err error

	cnt := 0
	for {
		u := uuid.NewV4()
		id := hex.EncodeToString(u[:])
		session := Session{id, internalID, time.Now().Unix() + int64(720*time.Hour)}

		_, err = dm.db.Insert(&session)

		if err == nil {
			return &id, nil
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
func (dm *DatabaseManager) SessionFind(sessionKey string) (*Session, error) {
	var resulsts []Session
	err := dm.db.Select(&resulsts, dm.db.Where("session_key", "=", sessionKey))

	if err != nil {
		return nil, err
	}

	if len(resulsts) == 0 {
		return nil, errors.New("Unknown session")
	}

	return &resulsts[0], nil
}

// SessionRemove is to remove session
// len(sessionKey) = 32
func (dm *DatabaseManager) SessionRemove(sessionKey string) error {
	_, err := dm.db.Delete(&Session{SessionKey: sessionKey})

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
