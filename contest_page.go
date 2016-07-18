package main

import "net/http"
import "github.com/naoina/genmai"
import "time"
import "text/template"
import "strconv"

type ContestsTopHandler struct {
    Temp *template.Template
    NumPerPage int
}

func CreateContestsTopHandler() (*ContestsTopHandler, error) {
    funcs := template.FuncMap{
        "TimeRangeToString": TimeRangeToString,
        "ContestTypeToString": func(t int64) string {
            return ContestTypeToString[ContestType(t)]
        },
    }

    temp, err := template.ParseFiles("./html/contests/index_tmpl.html")
    
    if err != nil {
        return nil, err
    }

    temp = temp.Funcs(funcs)

    return &ContestsTopHandler{temp, 50}, nil
}

func TimeRangeToString(start, finish int64) string {
    startTime := time.Unix(start, 0)
    finishTime := time.Unix(finish, 0)

    return startTime.Format("2006/01/02 15:04:05") + "-" + finishTime.Format("2006/01/02 15:04:05")
}

func (ch ContestsTopHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
    std, err := ParseRequestForSession(req)

	if std == nil || err != nil || !std.IsSignedIn {
		RespondRedirection(rw, "/contests")

        return
	}

    err = req.ParseForm()


    if err != nil {
        rw.WriteHeader(http.StatusBadRequest)

        rw.Write([]byte(BR400))

        return
    }

    var cond *genmai.Condition
    timeNow := time.Now().Unix()

    if req.URL.Path == "/" {
        cond = mainDB.db.Where("start_time", "<=", timeNow).And(mainDB.db.Where("finish_time", ">=", timeNow))
    }else if req.URL.Path == "/coming" {
        cond = mainDB.db.Where("start_time", ">", timeNow)
    }else if req.URL.Path == "/closed" {
        cond = mainDB.db.Where("finish_time", "<", timeNow)
    }else {
        // 各コンテストとコンテスト新規作成

        rw.WriteHeader(http.StatusNotImplemented)
        
        rw.Write([]byte(NI501))

        return
    }
    
    page := 0

    if queryArr, has := req.Form["p"]; has {
        p, err := strconv.ParseInt(queryArr[0], 10, 64)

        if err != nil || p <= 0 {
            page = 1
        }else {
            page = 1
        }
    }else {
        page = 1
    }

    count, err := mainDB.ContestCount(cond)

    if err != nil {
        //TODO 
        return
    }

    type TemplateVal struct {
        Contests []Contest
        UserName string
    }
    var templateVal TemplateVal
    templateVal.UserName = std.UserName

    if count > 0 {
        if (page - 1) * ch.NumPerPage + 1 > int(count) {
            page = 1
        }

        cond = cond.OrderBy("start_time", genmai.ASC).Offset(int((page - 1) * ch.NumPerPage + 1)).Limit(ch.NumPerPage)

        contests, err := mainDB.ContestList(cond)

        if err == nil {
            templateVal.Contests = *contests
        }
    }

    ch.Temp.Execute(rw, templateVal)

}


