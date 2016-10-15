package main

import (
	"bytes"
	"crypto/sha512"
	"errors"
	"fmt"
	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday"
	html "html/template"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"strings"
	"text/template"
	"time"
	"unicode/utf8"
)

var UTF8_BOM = []byte{239, 187, 191}
var UTF8_EDL = []byte{'\r', '\n'}

func HasBOM(in []byte) bool {
	return bytes.HasPrefix(in, UTF8_BOM)
}

func StripBOM(in []byte) []byte {
	return bytes.TrimPrefix(in, UTF8_BOM)
}

func StripEndline(in []byte) []byte {
	return bytes.TrimPrefix(in, UTF8_EDL)
}

func UTF8StringLengthAndBOMCheck(str string, l int) bool {
	if len(str) > l*6 {
		return false
	}

	if utf8.RuneCountInString(str) > l {
		return false
	}

	return !HasBOM([]byte(str))
}

func ReplaceEndline(str string) string {
	return strings.Replace(strings.Replace(str, "\r\n", "\n", -1), "\r", "\n", -1)
}

func TimeToString(t int64) string {
	return time.Unix(t, 0).Format("2006/01/02 15:04:05")
}

// CreateHandlers is a function to return hadlers
func CreateHandlers() (*map[string]*http.HandlerFunc, error) {
	res := make(map[string]*http.HandlerFunc)

	var err error

	res["/"], err = func() (*http.HandlerFunc, error) {
		funcs := template.FuncMap{
			"timeToString": TimeToString,
		}

		temp, err := template.New("").Funcs(funcs).ParseFiles("./html/index_tmpl.html")

		if err != nil {
			return nil, err
		}

		tmp := temp.Lookup("index_tmpl.html")

		if err != nil {
			return nil, errors.New("Failed to load ./html/index_tmpl.html")
		}

		f := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			if req.URL.Path != "/" && req.URL.Path != "/#" {
				rw.WriteHeader(http.StatusNotFound)

				fmt.Fprint(rw, NF404)

				return
			}

			std, err := ParseRequestForSession(req)

			if std == nil || err != nil {
				std = &SessionTemplateData{
					IsSignedIn: false,
					UserID:     "",
					UserName:   "",
				}
			}

			news, err := mainDB.NewsGet()

			if err != nil {
				news = make([]News, 0)

				fmt.Println(err.Error())
			}

			type IndexResp struct {
				*SessionTemplateData
				News      []News
				NewsCount int
			}

			resp := &IndexResp{std, news, mainDB.showedNewCount}

			rw.WriteHeader(http.StatusOK)
			tmp.Execute(rw, *resp)
		})

		return &f, nil
	}()

	if err != nil {
		return nil, err
	}

	res["/onlinejudge/"], err = func() (*http.HandlerFunc, error) {
		ojh := http.StripPrefix("/onlinejudge/", CreateOnlineJudgeHandler())

		if ojh == nil {
			return nil, errors.New("Failed to CreateOnlineJudgeHandler()")
		}

		f := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			rw.WriteHeader(http.StatusNotImplemented)

			fmt.Fprint(rw, NI501)

			return

			ojh.ServeHTTP(rw, req)
		})

		return &f, nil
	}()

	if err != nil {
		return nil, err
	}

	res["/contests/"], err = func() (*http.HandlerFunc, error) {
		contestsTopHandler, err := CreateContestsTopHandler()

		if err != nil {
			return nil, err
		}

		handler := http.StripPrefix("/contests", *contestsTopHandler)

		f := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			handler.ServeHTTP(rw, req)
		})

		return &f, nil
	}()

	if err != nil {
		return nil, err
	}

	res["/login"], err = func() (*http.HandlerFunc, error) {
		type LoginTemp struct {
			IsFailed bool
			BackURL  string
		}

		tmp, err := template.ParseFiles("./html/login_tmpl.html")

		if err != nil {
			return nil, errors.New("Failed to load ./html/login_tmpl.html")
		}

		f := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			if req.Method == "GET" {
				req.ParseForm()
				rw.WriteHeader(http.StatusOK)

				comeback, res := req.Form["comeback"]

				var cburl string
				if !res || len(comeback) == 0 || len(comeback[0]) == 0 {
					cburl = "/"
				} else {
					cburl = comeback[0]
				}

				tmp.Execute(rw, LoginTemp{false, cburl})
			} else if req.Method == "POST" {
				if err := req.ParseForm(); err != nil {
					rw.WriteHeader(http.StatusBadRequest)

					return
				}

				loginID, res := req.Form["loginID"]
				password, res2 := req.Form["password"]
				comeback, res3 := req.Form["comeback"]

				if !res || !res2 || !res3 || len(loginID) == 0 || len(password) == 0 || len(comeback) == 0 {
					rw.WriteHeader(http.StatusBadRequest)
					fmt.Fprint(rw, BR400)

					return
				}

				if strings.Index(comeback[0], "//") != -1 || len(loginID[0]) > 40 {
					rw.WriteHeader(http.StatusBadRequest)
					fmt.Fprint(rw, BR400)

					return
				}

				user, err := mainDB.UserFindFromUserID(loginID[0])
				passHash := sha512.Sum512([]byte(password[0]))

				if err != nil || !reflect.DeepEqual(user.PassHash, passHash[:]) {
					rw.WriteHeader(http.StatusOK)

					tmp.Execute(rw, LoginTemp{true, comeback[0]})

					return
				}

				sessionID, err := mainDB.SessionAdd(user.Iid)

				if err != nil {
					rw.WriteHeader(http.StatusInternalServerError)
					HttpLog.Println("page.go:261:", err)

					fmt.Fprint(rw, ISE500)
				} else {
					cookie := http.Cookie{
						Name:   "session",
						Value:  *sessionID,
						MaxAge: 2592000,
					}

					http.SetCookie(rw, &cookie)

					comeback[0] = /*"http://azure2.wt6.pw:10065" + */ comeback[0] // TODO: Load from setting file

					RespondRedirection(rw, comeback[0])
				}
			} else {
				rw.WriteHeader(http.StatusBadRequest)
				fmt.Fprint(rw, BR400)
			}

		})

		return &f, nil
	}()

	if err != nil {
		return nil, err
	}

	res["/logout"], err = func() (*http.HandlerFunc, error) {
		f := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			session := ParseSession(req)

			if session != nil {
				mainDB.SessionRemove(*session)
			}

			cookie := http.Cookie{
				Name:   "session",
				Value:  *session,
				MaxAge: 0,
			}

			http.SetCookie(rw, &cookie)
			RespondRedirection(rw, "/")
		})

		return &f, nil
	}()

	if err != nil {
		return nil, err
	}

	res["/userinfo"], err = func() (*http.HandlerFunc, error) {
		tmp, err := template.ParseFiles("./html/userinfo_tmpl.html")

		if err != nil {
			return nil, err
		}

		f := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			user, err := ParseRequestForUseData(req)

			if err != nil {
				RespondRedirection(rw, "/login?comeback=/userinfo")

				return
			}

			rw.WriteHeader(http.StatusOK)
			tmp.Execute(rw, user)
		})

		return &f, nil
	}()

	if err != nil {
		return nil, err
	}

	res["/userinfo/update_password"], err = func() (*http.HandlerFunc, error) {
		tmp, err := template.ParseFiles("./html/userinfo_update_password_tmpl.html")

		if err != nil {
			return nil, err
		}

		f := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			user, err := ParseRequestForUseData(req)

			if err != nil {
				RespondRedirection(rw, "/userinfo")

				return
			}

			rw.WriteHeader(http.StatusOK)
			tmp.Execute(rw, user)
		})

		return &f, nil
	}()

	if err != nil {
		return nil, err
	}

	res["/signup"], err = func() (*http.HandlerFunc, error) {
		_, err := template.ParseFiles("./html/signup_tmpl.html")

		if err != nil {
			return nil, err
		}

		f := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			rw.WriteHeader(http.StatusNotImplemented)
			rw.Write([]byte(NI501))

			return

			_, err := ParseRequestForUseData(req)

			if err != nil {
				RespondRedirection(rw, "/")

				return
			}

			type TeplateVal struct {
				ReCAPTCHA     bool
				Msg           string
				UserID        string
				UserName      string
				Email         string
				Password      string
				ReCAPTCHASite string
			}

			if req.Method == "GET" {

			}
		})

		return &f, nil
	}()

	res["/help"], err = func() (*http.HandlerFunc, error) {
		tmp, err := template.ParseFiles("./html/help_tmpl.html")

		if err != nil {
			return nil, err
		}

		type TemplateVal struct {
			Help       html.HTML
			UserName   string
			IsSignedIn bool
		}

		f := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			rw.WriteHeader(http.StatusNotFound)
			rw.Write([]byte(NF404))

			return

			std, err := ParseRequestForUseData(req)

			var IsSignedIn bool = false
			var Name string
			if err == nil {
				IsSignedIn = true
				Name = std.UserName
			}

			fp, err := os.Open("./html/help.md")

			if err != nil {
				HttpLog.Println(err)
				rw.WriteHeader(http.StatusInternalServerError)
				rw.Write([]byte(ISE500))

				return
			}
			defer fp.Close()

			b, err := ioutil.ReadAll(fp)

			if err != nil {
				HttpLog.Println(err)
				rw.WriteHeader(http.StatusInternalServerError)
				rw.Write([]byte(ISE500))

				return
			}

			unsafe := blackfriday.MarkdownCommon(b)
			page := bluemonday.UGCPolicy().SanitizeBytes(unsafe)

			tmp.Execute(rw, TemplateVal{html.HTML(string(page)), Name, IsSignedIn})
		})

		return &f, nil
	}()

	// Debug request
	res["/admin"], err = func() (*http.HandlerFunc, error) {
		tmp, err := template.ParseFiles("./html/admin_tmpl.html")

		if err != nil {
			return nil, err
		}

		f := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			err := req.ParseForm()

			if err != nil {
				rw.WriteHeader(http.StatusBadRequest)
				rw.Write([]byte(ISE500))

				return
			}

			std, err := ParseRequestForUseData(req)

			if err != nil || std.Gid != 0 {
				RespondRedirection(rw, "/")

				return
			}

			wrapFormStr := func(str string) string {
				arr, has := req.Form[str]
				if has && len(arr) != 0 {
					return arr[0]
				}
				return ""
			}

			/*type TeplateVal struct {
				ReCAPTCHA bool
				Msg string
				UserID string
				UserName string
				Email string
				Password string
				ReCAPTCHASite string
			}*/

			if req.Method == "GET" {
				tmp.Execute(rw, map[string]string{"UserName": std.UserName})
			} else {
				id := wrapFormStr("id")
				name := wrapFormStr("name")
				pass := wrapFormStr("password")

				if len(id) == 0 || len(name) == 0 || len(pass) == 0 {
					rw.WriteHeader(http.StatusBadRequest)
					rw.Write([]byte(BR400))

					return
				}

				_, err := mainDB.UserAdd(id, name, pass, id+"@hoge.com", 1)

				if err != nil {
					rw.WriteHeader(http.StatusBadRequest)
					rw.Write([]byte("BadRequest: " + err.Error()))
					return
				}

				RespondRedirection(rw, "/admin")
			}
		})

		return &f, nil
	}()

	if err != nil {
		return nil, err
	}

	return &res, nil
}
