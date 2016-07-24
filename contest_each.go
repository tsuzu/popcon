package main

import "net/http"
import "html/template"
import "strconv"
import "time"

type ContestEachHandler struct {
    Top *template.Template
    ProbList *template.Template
    ProbView *template.Template
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

    mux := http.NewServeMux()

    mux.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request){
        if req.URL.Path != "/" {
            rw.WriteHeader(http.StatusNotFound)

            rw.Write([]byte(NF404))

            return
        }

        type TemplateVal struct {
            UserName string
            Cid int64
            ContestName string
            Description template.HTML
            NotJoined bool
        }

        desc, err := mainDB.ContestDescriptionLoad(cid)

        if err != nil {
            desc = ""
        }

        joined := check
        if cont.FinishTime <= time.Now().Unix() {
            joined = true
        }

        templateVal := TemplateVal{
            UserName: std.UserName,
            Cid: cid,
            ContestName: cont.Name,
            Description: template.HTML(desc),
            NotJoined: !joined,
        }

        rw.WriteHeader(http.StatusOK)
        ceh.Top.Execute(rw, templateVal)
    })

    mux.HandleFunc("/problems/", func(rw http.ResponseWriter, req *http.Request) {
        if !check {
            RespondRedirection(rw, "/contests/" + strconv.FormatInt(cid, 10) + "/")

            return
        }

        http.StripPrefix("/problems/", http.HandlerFunc(func(rw http.ResponseWriter, req*http.Request){
            if req.URL.Path == "" {
                probList, err := mainDB.ContestProblemList(cid)

                if err != nil {
                    probList = &[]ContestProblem{}
                }

                type TemplateVal struct {
                    Problems []ContestProblem
                    UserName string
                    Cid int64
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
                Cid int64
                Text string
                UserName string
            }
            templateVal := TemplateVal{*prob, cid, *stat, std.UserName}

            rw.WriteHeader(http.StatusOK)

            ceh.ProbView.Execute(rw, templateVal)
        })).ServeHTTP(rw, req)
    })

    mux.HandleFunc("/join", func(rw http.ResponseWriter, req *http.Request){
        if req.Method == "GET" {
            mainDB.ContestParticipationAdd(std.Iid, cid)

            RespondRedirection(rw, "/contests/" + strconv.FormatInt(cid, 10) + "/")
        }else {
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

    return &ContestEachHandler{top, probList, probView}, nil
}