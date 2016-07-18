package main

import (
	"net/http"
	"time"
)

func main() {
	pmainDB, err := NewDatabaseManager()

	if err != nil {
		panic(err)
	}

/*	fmt.Println(pmainDB.UserAdd("tsuzu", "つづ", "hoge", "hoge@hoge.com", 0))
	fmt.Println(pmainDB.ContestNew("Hoge", time.Date(2016, 7, 18, 0, 0, 0, 0, time.FixedZone("Asia/Tokyo", 9*60*60)).Unix(), time.Date(2016, 7, 19, 22, 0, 0, 0, time.FixedZone("Asia/Tokyo", 9*60*60)).Unix(), 1, ContestJOI))

	for i := 0; i < 300; i++ {
		pmainDB.ContestNew("テキトーコンテスト" + fmt.Sprint(i), time.Date(2016, 7, 18, 0, 0, 0, 0, time.FixedZone("Asia/Tokyo", 9*60*60)).Unix(), time.Date(2016, 7, 19, 22, 0, 0, 0, time.FixedZone("Asia/Tokyo", 9*60*60)).Unix(), 1, ContestJOI)
	}*/

	// Copy to the global variable
	mainDB = pmainDB
	mainDB.showedNewCount = 5

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
