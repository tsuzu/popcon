package main

import "errors"

// SessionTemplateData contains data in SQL DB
type SessionTemplateData struct {
    IsSignedIn bool
    ID string
    ScreenName string
}

// GetSessionTemplateData returns a SessionTemplateData object
func GetSessionTemplateData(sessionID string) (*SessionTemplateData, error) {
    ID, err := mainDB.SessionFind(sessionID)
    
    if err != nil {
        return nil, err
    }
    
    if ID == nil {
        return nil, errors.New("session not found")
    }
    
    user, err := mainDB.UserFind(*ID)

    if err != nil || user == nil {
        return nil, errors.New("user not found")
    }
    
    return &SessionTemplateData{true, user.userID, user.userName}, nil
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