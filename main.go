package main

import (
	"encoding/json"
	"flag"
	"fmt"
	gorilla "github.com/gorilla/handlers"
	"github.com/sebest/xff"
	"net/http"
	"os"
	"golang.org/x/crypto/ssh/terminal"
	"syscall"
)

func CreateDefaultAdminUser() bool {
	fmt.Println("No user found in the DB")
	fmt.Println("You need to create the default admin")
	var id, name, email, pass, pass2 string

	fmt.Print("User ID: ")
	_, err := fmt.Scan(&id)

	if len(id) == 0 || err != nil {
		return false
	}

	fmt.Print("User Name: ")
	_, err = fmt.Scan(&name)

	if len(name) == 0 || err != nil {
		return false
	}

	fmt.Print("Email: ")
	_, err = fmt.Scan(&email)

	if len(email) == 0 || err != nil {
		return false
	}

	fmt.Print("Password (hidden): ")
	passArr, err := terminal.ReadPassword(int(syscall.Stdin))

	if err != nil {
		return false
	}

	fmt.Print("Password (confirmation): ")
	passArr2, err := terminal.ReadPassword(int(syscall.Stdin))

	if err != nil {
		return false
	}

	pass = string(passArr)
	pass2 = string(passArr2)

	if pass != pass2 {
		return false
	}

	_, err = mainDB.UserAdd(id, name, pass, email, 0)

	if err != nil {
		fmt.Println("Failed to create user. (", err.Error(), ")")

		return false
	}

	return true
}

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

	userCnt, err :=  mainDB.UserCount()

	if err != nil {
		DBLog.Println("Failed to count the users", err.Error())

		return
	}

	if userCnt == 0 {
		if !CreateDefaultAdminUser() {
			DBLog.Println("failed.")

			return
		}
	}

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
