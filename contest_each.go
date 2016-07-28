package main

import (
	"fmt"
	"text/template"
	htmlTemplate "html/template"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday"
)

type ContestEachHandler struct {
	Top      *template.Template
	ProbList *template.Template
	ProbView *template.Template
	SubList  *template.Template
	SubView  *template.Template
	Submit   *template.Template
	ManTop   *template.Template
	ManRej   *template.Template
	ManSet   *template.Template
	ManPro   *template.Template
	ManProV  *template.Template
}

func (ceh *ContestEachHandler) checkAdmin(cont *Contest, std SessionTemplateData) bool {
	if std.Gid == 0 {
		return true
	}

	if cont.Admin == std.Iid {
		return true
	}

	return false
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

	isStarted := (cont.StartTime >= time.Now().Unix())
	isFinished := (cont.FinishTime <= time.Now().Unix())

	free := (check && isStarted) || isFinished

	isAdmin := ceh.checkAdmin(cont, std)

	if isAdmin {
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
			UserName               string
			Cid                    int64
			ContestName            string
			Description            htmlTemplate.HTML
			JoinButtonActive       bool
			StartTime              int64
			FinishTime             int64
			Enabled                bool
			ManagementButtonActive bool
		}

		desc, err := mainDB.ContestDescriptionLoad(cid)

		if err != nil {
			desc = ""
		}

		unsafe := blackfriday.MarkdownCommon([]byte(desc))
		html := bluemonday.UGCPolicy().SanitizeBytes(unsafe)

		templateVal := TemplateVal{
			UserName:               std.UserName,
			Cid:                    cid,
			ContestName:            cont.Name,
			Description:            htmlTemplate.HTML(html),
			JoinButtonActive:       !(isFinished || check || isAdmin),
			StartTime:              cont.StartTime,
			FinishTime:             cont.FinishTime,
			Enabled:                free,
			ManagementButtonActive: isAdmin,
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
					ContestName string
					Problems    []ContestProblem
					UserName    string
					Cid         int64
				}

				templateVal := TemplateVal{
					cont.Name,
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

			unsafe := blackfriday.MarkdownCommon([]byte(*stat))
			html := bluemonday.UGCPolicy().SanitizeBytes(unsafe)

			type TemplateVal struct {
				ContestProblem
				ContestName string
				Cid         int64
				Text        string
				UserName    string
			}
			templateVal := TemplateVal{*prob, cont.Name, cid, string(html), std.UserName}

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
			} else {
				user, err := mainDB.UserFindFromUserID(userID)

				if err != nil {
					iid = IllegalParam
				} else {
					iid = user.Iid
				}
			}

			if !(isFinished || isAdmin) && iid != std.Iid {
				RespondRedirection(rw, "/contests/"+strconv.FormatInt(cid, 10)+"/submissions/?user="+std.UserID)

				return
			}

			count, err := mainDB.SubmissionViewCount(cid, iid, lang, prob, stat)

			if err != nil {
				fmt.Println(err) //TODO Fix

				return
			}

			type TemplateVal struct {
				AllEnabled  bool
				ContestName string
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
			templateVal.AllEnabled = isFinished || free
			templateVal.ContestName = cont.Name
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
			} else {
				templateVal.Languages = *langs
			}

			probs, err := mainDB.ContestProblemListLight(cid)

			if err != nil {
				fmt.Println(err)
			} else {
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
		} else {
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

			if submission.Cid != cid {
				rw.WriteHeader(http.StatusNotFound)

				rw.Write([]byte(NF404))

				return
			}

			if !isAdmin && submission.Iid != std.Iid {
				rw.WriteHeader(http.StatusForbidden)

				rw.Write([]byte(FBD403))

				return
			}

			code, err := mainDB.SubmissionGetCode(sid)

			if err != nil {
				var tmp string

				code = &tmp
			}

			casesMap, err := mainDB.SubmissionGetCase(sid)
			var cases []SubmissionTestCase

			if err == nil {
				cases = make([]SubmissionTestCase, len(*casesMap))
				idx := 0
				for _, v := range *casesMap {
					cases[idx] = v

					idx++
				}
			} else {
				fmt.Println(err) // TODO: Fix
			}

			msg := mainDB.SubmissionGetMsg(sid)

			if msg != nil && len(*msg) == 0 {
				msg = nil
			}

			type TemplateVal struct {
				ContestName string
				Submission  SubmissionViewEach
				Cases       []SubmissionTestCase
				Code        string
				Msg         *string
				UserName    string
				Cid         int64
			}

			templateVal := TemplateVal{
				ContestName: cont.Name,
				Submission:  *submission,
				Cases:       cases,
				Code:        *code,
				Msg:         msg,
				UserName:    std.UserName,
				Cid:         cid,
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
		if !free {
			RespondRedirection(rw, "/contests/"+strconv.FormatInt(cid, 10)+"/")

			return
		}

		if req.Method == "GET" {
			type TemplateVal struct {
				ContestName string
				UserName    string
				Cid         int64
				Problems    []ContestProblemLight
				Languages   []Language
				Prob        int64
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

			templateVal := TemplateVal{
				cont.Name,
				std.UserName,
				cid,
				*list,
				*lang,
				prob,
			}

			rw.WriteHeader(http.StatusOK)
			ceh.Submit.Execute(rw, templateVal)
		} else if req.Method == "POST" {
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

			lid := wrapForm("lang")
			pid := wrapForm("prob")
			code := wrapFormStr("code")

			if lid < 0 || pid < 0 || code == "" {
				rw.WriteHeader(http.StatusBadRequest)
				rw.Write([]byte(BR400))

				return
			}

			prob, err := mainDB.ContestProblemFind2(cid, pid)

			if err != nil {
				if err.Error() == "Unknown problem" {
					rw.WriteHeader(http.StatusBadRequest)
					rw.Write([]byte(BR400))

					return
				} else {
					rw.WriteHeader(http.StatusInternalServerError)
					rw.Write([]byte(ISE500))

					return
				}
			}

			_, err = mainDB.LanguageFind(lid)

			if err != nil {
				if err.Error() == "Unknown language" {
					rw.WriteHeader(http.StatusBadRequest)
					rw.Write([]byte(BR400))

					return
				} else {
					rw.WriteHeader(http.StatusInternalServerError)
					rw.Write([]byte(ISE500))

					return
				}
			}

			subm, err := mainDB.SubmissionNew(prob.Pid, std.Iid, lid, code)

			if err != nil {
				rw.WriteHeader(http.StatusInternalServerError)
				rw.Write([]byte(ISE500))

				return
			}

			RespondRedirection(rw, "/contests/"+strconv.FormatInt(cid, 10)+"/submissions/"+strconv.FormatInt(subm, 10))
		} else {

		}
	})

	mux.Handle("/management/", http.StripPrefix("/management/", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if !isAdmin {
			rw.WriteHeader(http.StatusForbidden)
			rw.Write([]byte(FBD403))

			return
		}

		if req.URL.Path == "" {
			type TemplateVal struct {
				Cid         int64
				UserName    string
				ContestName string
			}
			ceh.ManTop.Execute(rw, TemplateVal{cid, std.UserName, cont.Name})
		} else if req.URL.Path == "rejudge" {
			respondTemp := func(msg string) {
				type TemplateVal struct {
					Cid         int64
					UserName    string
					Msg         *string
					ContestName string
				}

				if msg == "" {
					ceh.ManRej.Execute(rw, TemplateVal{cid, std.UserName, nil, cont.Name})
				} else {
					ceh.ManRej.Execute(rw, TemplateVal{cid, std.UserName, &msg, cont.Name})
				}
			}

			if req.Method == "GET" {
				respondTemp("")

				return
			} else if req.Method == "POST" {
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

				target, id := wrapForm("target"), wrapForm("id")

				if (target != 1 && target != 2) || id < 0 {
					respondTemp("不正なIDです。")

					return
				}

				if target == 1 {
					_, err := mainDB.SubmissionFind(id)

					if err != nil {
						if err.Error() == "Unknown submission" {
							respondTemp("該当する提出がありません。")
						} else {
							rw.WriteHeader(http.StatusInternalServerError)
							rw.Write([]byte(ISE500))
						}
						return
					}
					// TODO: Rejudge処理を実装
					/*err = mainDB.SubmissionUpdate(id, 0, 0, InQueue, 0, 0)

					if err != nil {
						rw.WriteHeader(http.StatusInternalServerError)
						rw.Write([]byte(ISE500))

						return
					}*/

				} else {
					cp, err := mainDB.ContestProblemFind2(cid, id)

					if err != nil {
						if err.Error() == "Unknown problem" {
							respondTemp("該当する問題がありません。")
						} else {
							rw.WriteHeader(http.StatusInternalServerError)
							rw.Write([]byte(ISE500))
						}
						return
					}

					_, err = mainDB.SubmissionList(mainDB.db.Where("cid", "=", cid).And("pid", cp.Pid))

					if err != nil {
						rw.WriteHeader(http.StatusInternalServerError)
						rw.Write([]byte(ISE500))

						return
					}
				}

				RespondRedirection(rw, "/contests/"+strconv.FormatInt(cid, 10)+"/management/")

			} else {
				rw.WriteHeader(http.StatusBadRequest)
				rw.Write([]byte(BR400))

				return
			}
		} else if req.URL.Path == "setting" {
			type TemplateVal struct {
				Cid         int64
				UserName    string
				Msg         *string
				StartDate   string
				StartTime   string
				FinishDate  string
				FinishTime  string
				Description string
				ContestName string
			}

			if req.Method == "POST" {
				wrapFormStr := func(str string) string {
					arr, has := req.Form[str]
					if has && len(arr) != 0 {
						return arr[0]
					}
					return ""
				}
				startDate, startTime := wrapFormStr("start_date"), wrapFormStr("start_time")
				finishDate, finishTime := wrapFormStr("finish_date"), wrapFormStr("finish_time")
				description := wrapFormStr("description")
				contestName := wrapFormStr("contest_name")

				startStr := startDate + " " + startTime
				finishStr := finishDate + " " + finishTime

				if len(contestName) == 0 {
					msg := "コンテスト名が不正です。"
					templateVal := TemplateVal{
						cid, std.UserID, &msg, startDate, startTime, finishDate, finishTime, description, contestName,
					}

					ceh.ManSet.Execute(rw, templateVal)

					return
				}

				start, err := time.ParseInLocation("2006/01/02 15:04", startStr, Location)

				if err != nil {
					msg := "開始日時の値が不正です。"
					templateVal := TemplateVal{
						cid, std.UserID, &msg, startDate, startTime, finishDate, finishTime, description, contestName,
					}

					ceh.ManSet.Execute(rw, templateVal)

					return
				}

				finish, err := time.ParseInLocation("2006/01/02 15:04", finishStr, Location)

				if err != nil {
					msg := "終了日時の値が不正です。"
					templateVal := TemplateVal{
						cid, std.UserID, &msg, startDate, startTime, finishDate, finishTime, description, contestName,
					}

					ceh.ManSet.Execute(rw, templateVal)

					return
				}

				if start.Unix() >= finish.Unix() || start.Unix() < time.Now().Unix() {
					msg := "開始日時及び終了日時の値が不正です。"
					templateVal := TemplateVal{
						cid, std.UserID, &msg, startDate, startTime, finishDate, finishTime, description, contestName,
					}

					ceh.ManSet.Execute(rw, templateVal)

					return
				}

				err = mainDB.ContestUpdate(cid, contestName, start.Unix(), finish.Unix(), cont.Admin, 0)

				if err != nil {
					if strings.Index(err.Error(), "Duplicate") != -1 {
						msg := "すでに存在するコンテスト名です。"
						templateVal := TemplateVal{
							cid, std.UserID, &msg, startDate, startTime, finishDate, finishTime, description, contestName,
						}

						ceh.ManSet.Execute(rw, templateVal)

						return
					} else {
						rw.WriteHeader(http.StatusInternalServerError)
						rw.Write([]byte(ISE500))

						return
					}
				}

				err = mainDB.ContestDescriptionUpdate(cid, description)

				if err != nil {
					fmt.Println(err)
				}

				RespondRedirection(rw, "/contests/"+strconv.FormatInt(cid, 10)+"/management/")
			} else if req.Method == "GET" {
				desc, _ := mainDB.ContestDescriptionLoad(cid)

				templateVal := TemplateVal{
					Cid:         cid,
					UserName:    std.UserID,
					StartDate:   time.Unix(cont.StartTime, 0).Format("2006/01/02"),
					StartTime:   time.Unix(cont.StartTime, 0).Format("15:04"),
					FinishDate:  time.Unix(cont.FinishTime, 0).Format("2006/01/02"),
					FinishTime:  time.Unix(cont.FinishTime, 0).Format("15:04"),
					ContestName: cont.Name,
					Description: desc,
				}
				ceh.ManSet.Execute(rw, templateVal)
			} else {
				rw.WriteHeader(http.StatusBadRequest)
				rw.Write([]byte(BR400))
			}
		} else if len(req.URL.Path) >= 9 && req.URL.Path[:9] == "problems/" {
			http.StripPrefix("problems/", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				if req.URL.Path == "new" {
					type TemplateVal struct {
						Cid       int64
						ContestName string
						UserName  string
						Msg       *string
						Mode      bool
						Pidx      int64
						Name      string
						Time      int64
						Mem       int64
						Type      int64
						Prob      string
						Lang		  int64
						Languages []Language
						Code      string
					}

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

					languages, err := mainDB.LanguageList()

					if err != nil {
						rw.WriteHeader(http.StatusBadRequest)
						rw.Write([]byte(BR400))

						return
					}

					if req.Method == "GET" {
						ceh.ManPro.Execute(rw, TemplateVal{Cid: cid, ContestName: cont.Name, Time: 1, Mem: 32, UserName: std.UserName, Mode: true, Languages: *languages})

						return
					} else if req.Method == "POST" {
						pidx, name, time, mem := wrapForm("pidx"), wrapFormStr("problem_name"), wrapForm("time"), wrapForm("mem")
						jtype, prob, lid, code := wrapForm("type"), wrapFormStr("prob"), wrapForm("lang"), wrapFormStr("code")

						if pidx == -1 || time < 1 || time > 10 || mem < 32 || mem > 1024 || jtype < 0 || jtype > 1 || lid == -1 {
							rw.WriteHeader(http.StatusBadRequest)
							rw.Write([]byte(BR400))

							return
						}

						if _, err := mainDB.LanguageFind(lid); err != nil {
							if err.Error() == "Unknown language" {
								rw.WriteHeader(http.StatusBadRequest)
								rw.Write([]byte(BR400))

								return
							}else {
								rw.WriteHeader(http.StatusInternalServerError)
								rw.Write([]byte(ISE500))

								return
							}
						}

						cp, err := cont.ProblemAdd(pidx, name, time, mem, JudgeType(jtype))

						if err != nil {
							if strings.Index(err.Error(), "Duplicate") != -1 {
								msg := "使用されている問題番号です。"
								ceh.ManPro.Execute(rw, TemplateVal{cid, cont.Name, std.UserName, &msg, true, pidx, name, time, mem, jtype, prob, lid, *languages, code})

								return
							}else {
								rw.WriteHeader(http.StatusInternalServerError)
								rw.Write([]byte(ISE500))

								return
							}
						}

						err = cp.UpdateStatement(prob)

						if err != nil {
							rw.WriteHeader(http.StatusInternalServerError)
							rw.Write([]byte(ISE500))

							return
						}

						err = cp.UpdateChecker(lid, code)

						if err != nil {
							rw.WriteHeader(http.StatusInternalServerError)
							rw.Write([]byte(ISE500))

							return
						}

						RespondRedirection(rw, "/contests/" + strconv.FormatInt(cid, 10) + "/management/problems/" + strconv.FormatInt(cp.Pidx, 10))
					} else {
						rw.WriteHeader(http.StatusBadRequest)
						rw.Write([]byte(BR400))

						return
					}
				}else if req.URL.Path == "" {
					type TemplateVal struct {
						Cid int64
						ContestName string
						UserName string
						Problems []ContestProblem
					}

					list, err := mainDB.ContestProblemList(cid)

					if err != nil {
						rw.WriteHeader(http.StatusInternalServerError)
						rw.Write([]byte(ISE500))

						return
					}

					ceh.ManProV.Execute(rw, TemplateVal{cid, cont.Name, std.UserName, *list})
				}
			})).ServeHTTP(rw, req)
		} else {
			rw.WriteHeader(http.StatusNotFound)

			rw.Write([]byte(NF404))
		}
	})))

	handler := func(rw http.ResponseWriter, req *http.Request) {
		mux.ServeHTTP(rw, req)
	}

	return handler, nil
}

func CreateContestEachHandler() (*ContestEachHandler, error) {
	funcMap := template.FuncMap{
		"timeRangeToString": TimeRangeToString,
	}

	top, err := template.New("").Funcs(funcMap).ParseFiles("./html/contests/each/index_tmpl.html")

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

	funcMap = template.FuncMap{
		"timeToString": TimeToString,
		"add":          func(x, y int) int { return x + y },
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

	man, err := template.ParseFiles("./html/contests/each/management_tmpl.html")

	if err != nil {
		return nil, err
	}

	manre, err := template.ParseFiles("./html/contests/each/management/rejudge_tmpl.html")

	if err != nil {
		return nil, err
	}

	funcMap = template.FuncMap{
		"timeRangeToString": TimeRangeToString,
	}

	manse, err := template.New("").Funcs(funcMap).ParseFiles("./html/contests/each/management/setting_tmpl.html")

	if err != nil {
		return nil, err
	}

	manpr, err := template.ParseFiles("./html/contests/each/management/problem_set_tmpl.html")

	if err != nil {
		return nil, err
	}

	manprv, err := template.ParseFiles("./html/contests/each/management/problems_tmpl.html")

	if err != nil {
		return nil, err
	}

	return &ContestEachHandler{top.Lookup("index_tmpl.html"), probList, probView, subList.Lookup("submissions_tmpl.html"), subView.Lookup("submission_view_tmpl.html"), submit, man, manre, manse.Lookup("setting_tmpl.html"), manpr, manprv}, nil
}
