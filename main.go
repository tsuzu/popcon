package main

import (
	"encoding/json"
	"flag"
	"fmt"
	gorilla "github.com/gorilla/handlers"
	"github.com/sebest/xff"
	"net/http"
	"os"
)

func main() {
	settingFile := flag.String("setting", "./popcon.json", "the path to setting file")
	help := flag.Bool("help", false, "Show help")

	flag.Parse()

	if *help {
		flag.PrintDefaults()

		return
	}

	fp, err := os.OpenFile(*settingFile, os.O_RDONLY, 0664)

	if err != nil {
		b, _ := json.Marshal(Setting{})

		fmt.Println("Json Sample: ", string(b))

		return
	}

	dec := json.NewDecoder(fp)

	var setting Setting
	err = dec.Decode(&setting)

	if err != nil {
		fmt.Println("Syntax error: ", err)

		return
	}

	settingManager.Set(setting)

	lo, err := CreateLogOut()

	if err != nil {
		panic(err)
	}

	pmainDB, err := NewDatabaseManager()

	if err != nil {
		panic(err)
	}

	// Copy to the global variable
	mainDB = pmainDB
	mainDB.showedNewCount = 5
	SJQueue = CreateSubmissionJudgeQueue()
	CreateLogger(lo)
	go SJQueue.run()

	mux := http.NewServeMux()
	handlers, err := CreateHandlers()

	if err != nil {
		panic(err)
	}

	for k, v := range *handlers {
		mux.HandleFunc(k, v.PassHandler())
	}

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
	mux.Handle("/judge", JudgeTransfer{})

	xffh, err := xff.Default()

	if err != nil {
		panic(err)
	}

	logger := gorilla.LoggingHandler(lo, xffh.Handler(mux))

	xssProtector := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("X-XSS-Protection", "1")
		logger.ServeHTTP(rw, req)
	})

	// Should use TLS
	server := http.Server{
		Addr:           ":80",
		MaxHeaderBytes: 1 << 20,
		Handler:        xssProtector,
	}

	err = server.ListenAndServe()

	if err != nil {
		panic(err)
	}
}
