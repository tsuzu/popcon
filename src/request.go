package main

import "net/http"

// ParseRequestForSession returns SessionTemplateData
// Session is not found:  returns (nil, nil)
// Session is found:  returns (session, nil)
// An error occured: returns (nil, err)
func ParseRequestForSession(req *http.Request) (*SessionTemplateData, error) {
	cookies := req.Cookies()
	var session *string

	for idx := range cookies {
		if cookies[idx].Name == "session" {
			session = &cookies[idx].Value
		}
	}
    
    if session == nil {
        return nil, nil
    }
    
    return GetSessionTemplateData(*session)
}
