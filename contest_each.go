package main

import (
	"html/template"
	"net/http"
	"time"
)

import "strconv"

import "fmt"

type ContestEachHandler struct {
	Top      *template.Template
	ProbList *template.Template
	ProbView *template.Template
    SubList  *template.Template
	SubView  *template.Template
	Submit   *template.Template
}

func (ceh *ContestEachHandler) GetHandler(cid int64, std SessionTemplateData) (http.HandlerFunc, error) {
	cont, err := mainDB.ContestFind(cid)

	if err != nil {
		return nil, err
	}

	check, err := mainDB.ContestParticipationCheck(std.Iid, cid)

	if err != nil {
		return nil, err
	}

	free := check
	if cont.FinishTime <= time.Now().Unix() {
		free = true
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/" {
			rw.WriteHeader(http.StatusNotFound)

			rw.Write([]byte(NF404))

			return
		}

		type TemplateVal struct {
			UserName    string
			Cid         int64
			ContestName string
			Description template.HTML
			NotJoined   bool
		}

		desc, err := mainDB.ContestDescriptionLoad(cid)

		if err != nil {
			desc = ""
		}

		templateVal := TemplateVal{
			UserName:    std.UserName,
			Cid:         cid,
			ContestName: cont.Name,
			Description: template.HTML(desc),
			NotJoined:   !free,
		}

		rw.WriteHeader(http.StatusOK)
		ceh.Top.Execute(rw, templateVal)
	})

	mux.HandleFunc("/problems/", func(rw http.ResponseWriter, req *http.Request) {
		if !free {
			RespondRedirection(rw, "/contests/"+strconv.FormatInt(cid, 10)+"/")

			return
		}

		http.StripPrefix("/problems/", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			if req.URL.Path == "" {
				probList, err := mainDB.ContestProblemList(cid)

				if err != nil {
					probList = &[]ContestProblem{}
				}

				type TemplateVal struct {
					Problems []ContestProblem
					UserName string
					Cid      int64
				}

				templateVal := TemplateVal{
					*probList,
					std.UserName,
					cid,
				}

				rw.WriteHeader(http.StatusOK)
				ceh.ProbList.Execute(rw, templateVal)

				return
			}
			pidx, err := strconv.ParseInt(req.URL.Path, 10, 64)

			if err != nil {
				rw.WriteHeader(http.StatusNotFound)

				rw.Write([]byte(NF404))

				return
			}

			prob, err := mainDB.ContestProblemFind2(cid, pidx)

			if err != nil {
				rw.WriteHeader(http.StatusNotFound)

				rw.Write([]byte(NF404))

				return
			}

			stat, err := prob.LoadStatement()

			if err != nil {
				rw.WriteHeader(http.StatusInternalServerError)
				rw.Write([]byte(ISE500))

				return
			}

			type TemplateVal struct {
				ContestProblem
				Cid      int64
				Text     string
				UserName string
			}
			templateVal := TemplateVal{*prob, cid, *stat, std.UserName}

			rw.WriteHeader(http.StatusOK)

			ceh.ProbView.Execute(rw, templateVal)
		})).ServeHTTP(rw, req)
	})

	mux.Handle("/submissions/", http.StripPrefix("/submissions/", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if !free {
			RespondRedirection(rw, "/contests/"+strconv.FormatInt(cid, 10)+"/")

			return
		}

		if req.URL.Path == "" {
			wrapForm := func(str string) int64 {
				arr, has := req.Form[str]
				if has && len(arr) != 0 {
					val, err := strconv.ParseInt(arr[0], 10, 64)

					if err != nil {
						return -1
					} else {
						return val
					}
				}
				return -1
			}

			wrapFormStr := func(str string) string {
				arr, has := req.Form[str]
				if has && len(arr) != 0 {
					return arr[0]
				}
				return ""
			}

			stat := wrapForm("status")
			lang := wrapForm("lang")
			prob := wrapForm("prob")
			page := int(wrapForm("p"))
			userID := wrapFormStr("user")

			const IllegalParam = -128
			if page == -1 {
				page = 1
			}

            var iid int64
            if userID == "" {
                iid = -1
            }else {
    			user, err := mainDB.UserFindFromUserID(userID)

    			if err != nil {
	    			iid = IllegalParam
    			} else {
				    iid = user.Iid
			    }
            }

			count, err := mainDB.SubmissionViewCount(cid, iid, lang, prob, stat)

			if err != nil {
				fmt.Println(err) //TODO Fix

				return
			}

			type TemplateVal struct {
				UserName    string
				Cid         int64
                Uid         string
				Submissions []SubmissionView
                Problems    []ContestProblemLight
                Languages   []Language
				Current     int
				MaxPage     int
				Pagination  []PaginationHelper
				Lang        int64
				Prob        int64
				Status      int64
				User        string
			}
			var templateVal TemplateVal
			templateVal.Cid = cid
            templateVal.UserName = std.UserName
			templateVal.User = userID
			templateVal.Status = stat
			templateVal.Lang = lang
			templateVal.Prob = prob
            templateVal.Uid = std.UserID

            langs, err := mainDB.LanguageList()

            if err != nil {
                fmt.Println(err)
            }else {
                templateVal.Languages = *langs
            }

            probs, err := mainDB.ContestProblemListLight(cid)

            if err != nil {
                fmt.Println(err)
            }else {
                templateVal.Problems = *probs
            }

			templateVal.Current = 1

			templateVal.MaxPage = int(count) / ContentsPerPage

			if int(count)%ContentsPerPage != 0 {
				templateVal.MaxPage++
			} else if templateVal.MaxPage == 0 {
				templateVal.MaxPage = 1
			}

			if count > 0 {
				if (page-1)*ContentsPerPage > int(count) {
					page = 1
				}

				templateVal.Current = page

				submissions, err := mainDB.SubmissionViewList(cid, iid, lang, prob, stat, int64((page-1)*ContentsPerPage), ContentsPerPage)

				if err == nil {
					templateVal.Submissions = *submissions
				} else {
					fmt.Println(err)
				}
			}

            templateVal.Pagination = CreatePaginationHelper(templateVal.Current, templateVal.MaxPage, 3)

            rw.WriteHeader(200)

            ceh.SubList.Execute(rw, templateVal)
		}else {
			sid, err := strconv.ParseInt(req.URL.Path, 10, 64)

			if err != nil {
				rw.WriteHeader(http.StatusNotFound)
				rw.Write([]byte(NF404))

				return
			}

			submission, err := mainDB.SubmissionViewFind(sid)

			if err != nil {
				fmt.Println(err) // TODO: Fix

				rw.WriteHeader(http.StatusNotFound)
				rw.Write([]byte(NF404))

				return
			}

			code, err := mainDB.SubmissionGetCode(sid)

			if err != nil {
				var tmp string 

				code = &tmp
			}

			casesMap, err := mainDB.SubmissionGetCase(sid)
			var cases []SubmissionTestCase

			if err == nil  {
				cases = make([]SubmissionTestCase, len(*casesMap))
				idx := 0
				for _, v := range *casesMap {
					cases[idx] = v

					idx++
				}
			}else {
				fmt.Println(err) // TODO: Fix
			}

			msg := mainDB.SubmissionGetMsg(sid)

			if msg != nil && len(*msg) == 0 {
				msg = nil
			}

			type TemplateVal struct {
				Submission SubmissionViewEach
				Cases []SubmissionTestCase
				Code string
				Msg *string
				UserName string
				Cid int64
			}

			templateVal := TemplateVal {
				Submission: *submission,
				Cases: cases,
				Code: *code,
				Msg: msg,
				UserName: std.UserName,
				Cid: cid,
			}

			rw.WriteHeader(http.StatusOK)
			ceh.SubView.Execute(rw, templateVal)
		}
	})))

	mux.HandleFunc("/join", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method == "GET" {
			mainDB.ContestParticipationAdd(std.Iid, cid)

			RespondRedirection(rw, "/contests/"+strconv.FormatInt(cid, 10)+"/")
		} else {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte(BR400))
		}
	})

	mux.HandleFunc("/submit", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method == "GET" {
			type TemplateVal struct {
				UserName string
				Cid int64
				Problems []ContestProblemLight
				Languages []Language
				Prob int64
			}

			list, err := mainDB.ContestProblemListLight(cid)

			if err != nil {
				list = &[]ContestProblemLight{}

				fmt.Println(err) // TODO: Fix
			}

			lang, err := mainDB.LanguageList()

			if err != nil {
				lang = &[]Language{}

				fmt.Println(err)
			}

			probArr, has := req.Form["prob"]
			var prob int64 = -1

			if has && len(probArr) != 0 {
				p, err := strconv.ParseInt(probArr[0], 10, 64)

				if err != nil {
					prob = -1
				}
				prob = p
			}

			templateVal := TemplateVal {
				std.UserName,
				cid,
				*list,
				*lang,
				prob,
			}

			rw.WriteHeader(http.StatusOK)
			ceh.Submit.Execute(rw, templateVal)
		} else {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte(BR400))
		}
	})

	handler := func(rw http.ResponseWriter, req *http.Request) {
		mux.ServeHTTP(rw, req)
	}

	return handler, nil
}

func CreateContestEachHandler() (*ContestEachHandler, error) {
	top, err := template.ParseFiles("./html/contests/each/index_tmpl.html")

	if err != nil {
		return nil, err
	}

	probList, err := template.ParseFiles("./html/contests/each/problems_tmpl.html")

	if err != nil {
		return nil, err
	}

	probView, err := template.ParseFiles("./html/contests/each/problem_view_tmpl.html")

	if err != nil {
		return nil, err
	}

    funcMap := template.FuncMap{
        "timeToString": TimeToString,
        "add" : func(x, y int) int {return x + y},
    }

    subList, err := template.New("").Funcs(funcMap).ParseFiles("./html/contests/each/submissions_tmpl.html")

	if err != nil {
		return nil, err
	}

	subView, err := template.New("").Funcs(funcMap).ParseFiles("./html/contests/each/submission_view_tmpl.html")

	if err != nil {
		return nil, err
	}

	submit, err := template.ParseFiles("./html/contests/each/submit_tmpl.html")

	if err != nil {
		return nil, err
	}

	return &ContestEachHandler{top, probList, probView, subList.Lookup("submissions_tmpl.html"), subView.Lookup("submission_view_tmpl.html"), submit}, nil
}
