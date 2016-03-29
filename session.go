package main

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

	return &SessionTemplateData{true, user.userID, user.userName}, nil
}

// GetSessionUserData returns an User object
func GetSessionUserData(sessionID string) (*User, error) {
	internalID, err := mainDB.SessionFind(sessionID)

	if err != nil {
		return nil, err
	}

	return mainDB.UserFindFromInternalID(internalID)
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
