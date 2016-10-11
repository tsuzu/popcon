package main

// Not Implemented

import "net/http"

// OnlineJudgeHandler is for /onlinejudge
type OnlineJudgeHandler struct {
}

// Problems is for  /onlinejudge/problems
func (ojk *OnlineJudgeHandler) Problems(rw http.ResponseWriter, req *http.Request) {
	if req.URL.Path == "" {

	}
}

// CreateOnlineJudgeHandler returns a handler for /onlinejudge
func CreateOnlineJudgeHandler() *http.ServeMux {
	ojh := OnlineJudgeHandler{}
	mux := http.NewServeMux()

	mux.Handle("/problems/", http.StripPrefix("/problems/", http.HandlerFunc(ojh.Problems)))

	return mux
}
