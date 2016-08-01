package main

import "github.com/gorilla/websocket"
import "net/http"
import "strings"
import "strconv"

type JudgeRequest struct {
	Sid         int64 // Submission ID
	Code        string
	Lang        int64
	Type        JudgeType
	Checker     string
	CheckerLang int64
	Cases       map[string]TestCase
	Time        int64
	Mem         int64
}

type JudgeResponse struct {
	Sid      int64 //SubmissionID
	Status   SubmissionStatus
	Msg      string
	Time     int64
	Mem      int64
	Case     int
	CaseName string
}

type JudgeTransfer struct {
}

func (jt JudgeTransfer) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	auth, has := req.Header["Authentication"]

	if !has || len(auth) == 0 || auth[0] != settings.JudgeKey {
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte(BR400))

		return
	}

	upgrader := websocket.Upgrader{}

	c, err := upgrader.Upgrade(rw, req, nil)

	if err != nil {
		return
	}

	defer c.Close()

	type TransferedResponse struct {
		Resp     JudgeResponse
		NewJudge int
	}

	for {
		var tr TransferedResponse
		err := c.ReadJSON(&tr)

		if err != nil {
			DBLog.Println(err)
			break
		}

        		go func(nj int) {
			for i := 0; i < nj; {
				sid := SJQueue.Pop()

				subm, err := mainDB.SubmissionFind(sid)

				if err != nil {
					DBLog.Println(err)

					continue
				}

				code, err := mainDB.SubmissionGetCode(sid)

				if err != nil {
					DBLog.Println(err)

					continue
				}

				cp, err := mainDB.ContestProblemFind(subm.Pid)

				if err != nil {
					DBLog.Println(err)

					continue
				}

				var checker string
				var checkerLang int64

				if cp.Type == int(JudgeRunningCode) {
					checkerLang, checker, err = cp.LoadChecker()

					if err != nil {
						DBLog.Println(err)

						continue
					}
				}

				cases, _, err := cp.LoadTestCases()

				if err != nil {
					DBLog.Println(err)

					continue
				}

				casesm := make(map[string]TestCase)

				for i := range *cases {
					casesm[strconv.FormatInt(int64(i), 32)] = (*cases)[i]
				}

				jr := JudgeRequest{
					Sid:         sid,
					Code:        *code,
					Lang:        subm.Lang,
					Type:        JudgeType(cp.Type),
					Checker:     checker,
					CheckerLang: checkerLang,
					Time:        cp.Time,
					Mem:         cp.Mem,
					Cases:       casesm,
				}

				err = c.WriteJSON(jr)

				if err != nil {
					DBLog.Println(err)

					SJQueue.Push(sid)

					break
				}

				i++
			}
		}(tr.NewJudge)

		if tr.Resp.Sid != -1 {
			if tr.Resp.Case != -1 {
				if tr.Resp.Status == Judging {
					arr := strings.Split(tr.Resp.Msg, "/")

					if len(arr) == 2 {
						fin, _ := strconv.ParseInt(arr[0], 10, 32)
						all, _ := strconv.ParseInt(arr[1], 10, 32)

						err = mainDB.SubmissionUpdate(tr.Resp.Sid, 0, 0, Judging, int(fin), int(all), 0)
						if err != nil {
							DBLog.Println(err)
						}
					}
				} else {
					if err == nil {
						if tr.Resp.Case == 0 {
							err = mainDB.SubmissionClearCase(tr.Resp.Sid)
						}

						if err != nil {
							DBLog.Println(err)
						} else {

							err := mainDB.SubmissionSetCase(tr.Resp.Sid, tr.Resp.Case, SubmissionTestCase{
								tr.Resp.Status,
								tr.Resp.CaseName,
								tr.Resp.Time,
								tr.Resp.Mem,
							})

							if err != nil {
								DBLog.Println(err)
							}
						}
					} else {
						DBLog.Println(err)
					}
				}
			} else {
                sm, err := mainDB.SubmissionFind(tr.Resp.Sid)

                if err != nil {
                    DBLog.Println(err)

                    return
                }

                cases, err := mainDB.SubmissionGetCase(tr.Resp.Sid)
                score := 0

                if err != nil {
                    DBLog.Println(err)

                    return
                }
                _, sets, err := (&ContestProblem{Pid: sm.Pid}).LoadTestCases()

                if err != nil {
                    DBLog.Println(err)
                }

                for i := range *sets {
                    ac := true

                    for i := range (*sets)[i].Cases {
                        c := (*sets)[i].Cases[i]

                        if tc, has := (*cases)[c]; !has || tc.Status != Accepted {
                            ac = false
                        }
                    }

                    if ac {
                        score += (*sets)[i].Score
                    }
                }

				err = mainDB.SubmissionUpdate(tr.Resp.Sid, tr.Resp.Time, tr.Resp.Mem, tr.Resp.Status, 0, 0, int64(score))

				if err != nil {
					DBLog.Println(err)
				}

				SJQueue.Remove(tr.Resp.Sid)

				mainDB.SubmissionSetMsg(tr.Resp.Sid, tr.Resp.Msg)

				if err != nil {
					DBLog.Println(err)
				}
			}
		}
	}
}
