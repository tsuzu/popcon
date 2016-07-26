package main

import "net/http"
import "github.com/naoina/genmai"
import "time"
import "text/template"
import "strconv"
import "fmt"
import "errors"
import "net/url"
import "strings"

type ContestsTopHandler struct {
    Temp *template.Template
    NumPerPage int
    EachHandler *ContestEachHandler
}

func CreateContestsTopHandler() (*ContestsTopHandler, error) {
    funcs := template.FuncMap{
        "add": func(x, y int) int {return x + y},
        "timeRangeToString": TimeRangeToString,
        "contestTypeToString": func(t int64) string {
            return ContestTypeToString[ContestType(t)]
        },
    }

    temp, err := template.New("").Funcs(funcs).ParseFiles("./html/contests/index_tmpl.html")
    
    if err != nil {
        return nil, err
    }

    temp = temp.Lookup("index_tmpl.html")

    if temp == nil {
        return nil, errors.New("Failed to load /contests/index_temp.html")
    }

    ceh, err := CreateContestEachHandler()

    if err != nil {
        return nil, err
    }

    return &ContestsTopHandler{temp, 50, ceh}, nil
}

func TimeRangeToString(start, finish int64) string {
    startTime := time.Unix(start, 0)
    finishTime := time.Unix(finish, 0)

    return startTime.Format("2006/01/02 15:04:05") + "-" + finishTime.Format("2006/01/02 15:04:05")
}

func (ch ContestsTopHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
    std, err := ParseRequestForSession(req)

	if std == nil || err != nil || !std.IsSignedIn {
		RespondRedirection(rw, "/login?comeback=/contests" + url.QueryEscape(req.URL.Path))

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
    var reqType int

    switch req.URL.Path {
        case "/":
            reqType = 0
            cond = mainDB.db.Where("start_time", "<=", timeNow).And(mainDB.db.Where("finish_time", ">", timeNow))
        case "/coming/":
            reqType = 1
            cond = mainDB.db.Where("start_time", ">", timeNow)
        case "/closed/":
            reqType = 2
            cond = mainDB.db.Where("finish_time", "<=", timeNow)
        case "/new/":
            rw.WriteHeader(http.StatusNotImplemented)
        
            rw.Write([]byte(NI501))
        default:
            if len(req.URL.Path) == 0 {
                RespondRedirection(rw, "/contests/")

                return
            }

            idx := strings.Index(req.URL.Path[1:], "/")

            if idx == -1 {
                RespondRedirection(rw, "/contests" + req.URL.Path + "/")

                return
            }

            cidStr := req.URL.Path[1:][:idx]

            cid, err := strconv.ParseInt(cidStr, 10, 64)

            if err != nil {
                rw.WriteHeader(http.StatusNotFound)
                rw.Write([]byte(NF404))

                return
            }

            handler, err := ch.EachHandler.GetHandler(cid, *std)

            if err != nil {
                rw.WriteHeader(http.StatusNotFound)
                rw.Write([]byte(NF404))

                return
            }

            http.StripPrefix("/" + cidStr, handler).ServeHTTP(rw, req)
            
            return
    }

    page := 1

    if queryArr, has := req.Form["p"]; has {
        p, err := strconv.ParseInt(queryArr[0], 10, 64)

        if err != nil || p <= 0 {
            page = 1
        }else {
            page = int(p)
        }
    }

    count, err := mainDB.ContestCount(cond)

    if err != nil {
        //TODO 
        fmt.Println(err)
        return
    }

    type TemplateVal struct {
        Contests []Contest
        UserName string
        Type int
        Current int
        MaxPage int
        Pagination []PaginationHelper
    }
    var templateVal TemplateVal
    templateVal.UserName = std.UserName
    templateVal.Type = reqType
    templateVal.Current = 1

    templateVal.MaxPage = int(count) / ch.NumPerPage

    if int(count) % ch.NumPerPage != 0 {
        templateVal.MaxPage++
    }else if templateVal.MaxPage == 0 {
        templateVal.MaxPage = 1
    }

    if count > 0 {
        if (page - 1) * ch.NumPerPage > int(count) {
            page = 1
        }

        templateVal.Current = page

        contests, err := mainDB.ContestList(cond, mainDB.db.OrderBy("start_time", genmai.ASC).Offset(int((page - 1) * ch.NumPerPage)).Limit(ch.NumPerPage))

        if err == nil {
            templateVal.Contests = *contests
        }else {
            fmt.Println(err)
        }
    }

    templateVal.Pagination = CreatePaginationHelper(templateVal.Current, templateVal.MaxPage, 3)

    rw.WriteHeader(http.StatusOK)
    ch.Temp.Execute(rw, templateVal)

}