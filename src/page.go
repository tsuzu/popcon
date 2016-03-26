package main

import (
	"net/http"
	"text/template"
    "fmt"
    "crypto/sha512"
)

// NF404 is "404 Not Found"
const NF404 = `
<!DOCTYPE html>
<html>
<head>
<title>
404 Not Found
</title>
</head>
<body>
<h1>404 Not Found</h1>
The page is not found in this server.
</body>
</html>
`

// PageError is a error type for Page.go
type PageError string

func (pe PageError) Error() string {
	return string(pe)
}

// NewPageError is to return a PageError
func NewPageError(msg string) PageError {
	return PageError(msg)
}

// PageHandlerFuncType is a type of the function used in PageHandler
type PageHandlerFuncType func(*template.Template, http.ResponseWriter, *http.Request)

// PageHandler a handler for each page
type PageHandler struct {
	ParsedPage *template.Template
	Callback   PageHandlerFuncType
}

// PassHandler is a function that pass a handler
func (ph *PageHandler) PassHandler() func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		ph.Callback(ph.ParsedPage, rw, req)
	}
}

// CreateHandlers is a function to return hadlers
func CreateHandlers() (*map[string]*PageHandler, error) {
	res := make(map[string]*PageHandler)

	var err error

	res["/"], err = func() (*PageHandler, error) {
		tmpl, err1 := template.ParseFiles("./html/index_tmpl.html")

		if err1 != nil {
			return nil, NewPageError("Failed to load ./html/index_tmpl.html")
		}

		f := func(tmp *template.Template, rw http.ResponseWriter, req *http.Request) {
			rw.WriteHeader(http.StatusOK)

			cookies := req.Cookies()
			var session *string
            
			for idx := range cookies {
				if cookies[idx].Name == "session" {
					session = &cookies[idx].Value
				}
			}
            
			var userName *string
			var userID *string
			var user *User
			if session != nil {
				userID, err = mainDB.SessionFind(*session)
                
				if userID != nil && err == nil {
					user, err1 = mainDB.UserFind(*userID)
                    
					if user != nil && err1 == nil {
						userName = &user.userName
					}
				}
			}

            var ID, UN string
            if userID != nil {
                ID = *userID
            }
            if userName != nil {
                UN = *userName
            }
            
			type Toppage struct {
				IsSignedIn bool
				ID         string
				ScreenName string
			}

			tmp.Execute(rw, Toppage{userID != nil && userName != nil, ID, UN})
		}

		return &PageHandler{tmpl, f}, nil
	}()

	if err != nil {
		return nil, err
	}

	res["/onlinejudge"], err = func() (*PageHandler, error) {
		tmpl, err1 := template.ParseFiles("./html/index_tmpl.html")

		if err1 != nil {
			return nil, NewPageError("Failed to load ./html/index_tmpl.html")
		}

		f := func(tmp *template.Template, rw http.ResponseWriter, req *http.Request) {
			rw.WriteHeader(http.StatusOK)

			type Toppage struct {
				IsSignedIn bool
				ID         string
				ScreenName string
			}

			tmp.Execute(rw, Toppage{false, "Tsuzu", "Tsuzu"})
		}

		return &PageHandler{tmpl, f}, nil
	}()

	res["/login"], err = func() (*PageHandler, error) {
		tmpl, err1 := template.ParseFiles("./html/login_tmpl.html")

		if err1 != nil {
			return nil, NewPageError("Failed to load ./html/login_tmpl.html")
		}

		f := func(tmp *template.Template, rw http.ResponseWriter, req *http.Request) {
            if req.Method == "GET" {
                rw.WriteHeader(http.StatusOK)

    			tmp.Execute(rw, nil)
            }else if req.Method == "POST" {
                if req.ParseForm() != nil {
                    rw.WriteHeader(http.StatusBadRequest)
                    rw.Write(nil)
                    
                    return
                }
                
                var str string
                
                fmt.Fscan(req.Body, str)
                
                loginID, res := req.Form["loginID"]
                password, res2 := req.Form["password"]
                
                if !res || !res2 || len(loginID) == 0 || len(password) == 0 || len(loginID) > 20 {
                    rw.WriteHeader(http.StatusBadRequest)
                    rw.Write(nil)
                    
                    return
                }
                
                user, err := mainDB.UserFind(loginID[0])
                
                if err != nil || user.passHash != sha512.Sum512([]byte(password[0])) {
                    rw.WriteHeader(http.StatusOK)
                    
                    tmp.Execute(rw, nil)
                    
                    return
                }
                
                sessionID, e := mainDB.SessionAdd(user.userID)
                
                if e != nil {
                    rw.WriteHeader(http.StatusInternalServerError)
                    
                    fmt.Fprint(rw, "<html><body><h1>500 Internal Server Error</h1></body></html>")
                }else {
                    cookie := http.Cookie{
                        Name: "session",
                        Value: *sessionID,
                    };
                    
                    cookie.MaxAge = 2592000
                    
                    http.SetCookie(rw, &cookie)
                    rw.Header().Add("Location", "/")
                    rw.WriteHeader(http.StatusFound)

                    rw.Write(nil)
                }
            }else {
                rw.WriteHeader(http.StatusBadRequest)                
                rw.Write(nil)
            }
			
		}

		return &PageHandler{tmpl, f}, nil
	}()

	if err != nil {
		return nil, err
	}

	return &res, nil
}
