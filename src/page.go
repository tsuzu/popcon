package main

import (
	"net/http"
	"text/template"
    "fmt"
    "crypto/sha512"
	"errors"
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

// NI501 is "501 Not Implemented"
const NI501 = `
<!DOCTYPE html>
<html>
<head>
<title>
501 Not Implemented
</title>
</head>
<body>
<h1>501 Not Implemented</h1>
The service is not implemented.
</body>
</html>
`

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
			return nil, errors.New("Failed to load ./html/index_tmpl.html")
		}

		f := func(tmp *template.Template, rw http.ResponseWriter, req *http.Request) {
			if req.URL.Path != "/" && req.URL.Path != "/#" {
				rw.WriteHeader(http.StatusNotFound)
				
				fmt.Fprint(rw, NF404)
				
				return
			}
			
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

            var ID, ScreenName string
            if userID != nil {
                ID = *userID
            }
            if userName != nil {
                ScreenName = *userName
            }
            
			type Toppage struct {
				IsSignedIn bool
				ID         string
				ScreenName string
			}

			tmp.Execute(rw, Toppage{userID != nil && userName != nil, ID, ScreenName})
		}

		return &PageHandler{tmpl, f}, nil
	}()

	if err != nil {
		return nil, err
	}

	res["/onlinejudge/"], err = func() (*PageHandler, error) {
		f := func(tmp *template.Template, rw http.ResponseWriter, req *http.Request) {
			rw.WriteHeader(http.StatusNotImplemented)
			
			fmt.Fprint(rw, NI501)
		}

		return &PageHandler{nil, f}, nil
	}()
	
	res["/contests/"], err = func() (*PageHandler, error) {
		f := func(tmp *template.Template, rw http.ResponseWriter, req *http.Request) {
			rw.WriteHeader(http.StatusNotImplemented)
			
			fmt.Fprint(rw, NI501)
		}

		return &PageHandler{nil, f}, nil
	}()

	res["/login"], err = func() (*PageHandler, error) {
		type LoginTemp struct {
			IsFailed bool
			BackURL string
		}
		
		tmpl, err1 := template.ParseFiles("./html/login_tmpl.html")

		if err1 != nil {
			return nil, errors.New("Failed to load ./html/login_tmpl.html")
		}

		f := func(tmp *template.Template, rw http.ResponseWriter, req *http.Request) {
            if req.Method == "GET" {
				req.ParseForm()
                rw.WriteHeader(http.StatusOK)
				
				comeback, res := req.Form["comeback"]

				var cburl string
				if !res || len(comeback) == 0 || len(comeback[0]) == 0 {
					cburl = "/"
				}else {
					cburl = comeback[0]
				}

    			tmp.Execute(rw, LoginTemp{false, cburl})
            }else if req.Method == "POST" {
                if req.ParseForm() != nil {
                    rw.WriteHeader(http.StatusBadRequest)
                    rw.Write(nil)
                    
                    return
                }
                
                loginID, res := req.Form["loginID"]
                password, res2 := req.Form["password"]
				backurl, res3 := req.Form["comeback"]
				
                if !res || !res2 || !res3 || len(loginID) == 0 || len(password) == 0 || len(backurl) == 0 || len(loginID) > 20 || len(backurl[0]) == 0 {
                    rw.WriteHeader(http.StatusBadRequest)
                    rw.Write(nil)
                    
                    return
                }
                
				if backurl[0][0] != '/' {
					rw.WriteHeader(http.StatusBadRequest)
                    rw.Write(nil)
                    
                    return
				}
				
                user, err := mainDB.UserFind(loginID[0])
                
                if err != nil || user.passHash != sha512.Sum512([]byte(password[0])) {
                    rw.WriteHeader(http.StatusOK)
                    
                    tmp.Execute(rw, LoginTemp{true, backurl[0]})
                    
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
                    rw.Header().Add("Location", backurl[0])
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
