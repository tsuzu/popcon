package main

import (
	"net/http"
	gorilla "github.com/gorilla/handlers"
	"github.com/sebest/xff"
	"flag"
	"os"
	"encoding/json"
	"fmt"
)

type Settings struct {
	ReCAPTCHASite string
	ReCAPTCHASecret string
	AddUser bool
	CreateContest bool
	DB string
	JudgeKey string
}
var settings Settings

func main() {
	setting := flag.String("setting", "./popcon.json", "the path to setting file")
	help := flag.Bool("help", false, "Show help")

	flag.Parse()

	if *help {
		flag.PrintDefaults()

		return
	}

	fp, err := os.OpenFile(*setting, os.O_RDONLY, 0664)

	if err != nil {
		b, _ := json.Marshal(Settings{})

		fmt.Println("Json Sample: ", string(b))

		return
	}

	dec := json.NewDecoder(fp)

	err = dec.Decode(&settings)

	if err != nil {
		fmt.Println("Syntax error: ", err)

		return
	}

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
		Addr:           ":8080",
		MaxHeaderBytes: 1 << 20,
		Handler:        xssProtector,
	}

	err = server.ListenAndServe()

	if err != nil {
		panic(err)
	}
}
