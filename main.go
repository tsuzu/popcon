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
