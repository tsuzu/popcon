package main

import (
	"net/http"
	"time"
	//"fmt"
)

func main() {
	pmainDB, err := NewDatabaseManager()

	if err != nil {
		panic(err)
	}

	// Copy to the global variable
	mainDB = pmainDB
	mainDB.showedNewCount = 5

	/*
	fmt.Println(pmainDB.ContestNew("Hoge", time.Date(2016, 7, 18, 0, 0, 0, 0, time.FixedZone("Asia/Tokyo", 9*60*60)).Unix(), time.Date(2016, 7, 19, 22, 0, 0, 0, time.FixedZone("Asia/Tokyo", 9*60*60)).Unix(), 1, ContestJOI))

	
	cid, er := pmainDB.ContestNew("サンプルコン！4", time.Now().Unix() - 100000, time.Now().Unix() + 100000, 1, ContestJOI)

	if er != nil {
		panic(er)
	}

	fmt.Println(cid)

	cont, er := pmainDB.ContestFind(cid)

	if er != nil {
		panic(er)
	}

	prob, er := cont.ProblemAdd(1, "Hello, world!", 2, 128, JudgePerfectMatch)

	if er != nil {
		panic(er)
	}

	prob.UpdateStatement("Hello, world!を出力するプログラムを作成せよ。")

	fmt.Println(*prob)*/

	//mainDB.NewsAdd("その点トッポってすげぇよな、最後までチョコたっぷりだもん。")
	//fmt.Println(mainDB.LanguageAdd("C++14", ""))
	/*for i := 0; i < 300; i++ {
		mainDB.SubmissionNew(5, 1, 1, "Hello, world!")
	}*/

	/*pmainDB.UserAdd("tsuzu", "つづ", "hoge", "hoge@hoge.com", 0)
	/*fmt.Println(pmainDB.ContestNew("Hoge", time.Date(2016, 7, 18, 0, 0, 0, 0, time.FixedZone("Asia/Tokyo", 9*60*60)).Unix(), time.Date(2016, 7, 19, 22, 0, 0, 0, time.FixedZone("Asia/Tokyo", 9*60*60)).Unix(), 1, ContestJOI))

	
	cid, er := pmainDB.ContestNew("サンプルコン！", time.Now().Unix() - 100000, time.Now().Unix() + 100000, 1, ContestJOI)

	if er != nil {
		panic(er)
	}

	cont, er := pmainDB.ContestFind(cid)

	if er != nil {
		panic(er)
	}

	prob, er := cont.ProblemAdd(1, "Hello, world!", 2, 128, JudgePerfectMatch)

	if er != nil {
		panic(er)
	}

	prob.UpdateStatement("Hello, world!を出力するプログラムを作成せよ。")

	fmt.Println(*prob)
*/

	mux := http.NewServeMux()
	handlers, err := CreateHandlers()

	if err != nil {
		panic(err)
	}

	for k, v := range *handlers {
		mux.HandleFunc(k, v.PassHandler())
	}

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

	// Should use TLS
	server := http.Server{
		Addr:           ":8080",
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
		Handler:        mux,
	}

	err = server.ListenAndServe()

	if err != nil {
		panic(err)
	}
}
