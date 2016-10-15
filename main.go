package main

import (
	"encoding/json"
	"flag"
	"fmt"
	gorilla "github.com/gorilla/handlers"
	"github.com/sebest/xff"
	"golang.org/x/crypto/ssh/terminal"
	"net/http"
	"os"
	"syscall"
	_ "net/http/pprof"
	"net"
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
	fmt.Println()

	fmt.Print("Password (confirmation): ")
	passArr2, err := terminal.ReadPassword(int(syscall.Stdin))

	if err != nil {
		return false
	}
	fmt.Println()

	pass = string(passArr)
	pass2 = string(passArr2)

	if pass != pass2 {
		fmt.Println("Different password")

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
	pprof := flag.String("pprof-port", "", "To run pprof server, set the address to listen")
	help := flag.Bool("help", false, "Show help")

	flag.Parse()

	if *help {
		flag.PrintDefaults()

		return
	}

	if len(*pprof) != 0 {
		l, err := net.Listen("tcp", *pprof)
    	
		if err != nil {
    	    panic(err)
	    }
    	fmt.Printf("pprof server is listening on %s\n", l.Addr())
	    go http.Serve(l, nil)
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

	userCnt, err := mainDB.UserCount()

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
		HttpLog.Fatal(err)
	}

	for k, v := range *handlers {
		mux.HandleFunc(k, *v)
	}

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
	mux.Handle("/judge", JudgeTransfer{})
	mux.HandleFunc("/favicon.ico", func() http.HandlerFunc {
		faviconPath := settingManager.Get().FaviconPath
		return func(rw http.ResponseWriter, req *http.Request) {
			if len(faviconPath) == 0 {
				rw.WriteHeader(http.StatusNotFound)
				rw.Write([]byte(NF404))
			}else {
				http.ServeFile(rw, req, faviconPath)
			}
		}
	}())

	xffh, err := xff.Default()

	if err != nil {
		HttpLog.Fatal(err)
	}

	logger := gorilla.LoggingHandler(lo, xffh.Handler(mux))

	xssProtector := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("X-XSS-Protection", "1")
		logger.ServeHTTP(rw, req)
	})

	// Should use TLS
	server := http.Server{
		Addr:           settingManager.Get().ListeningEndpoint,
		MaxHeaderBytes: 1 << 20,
		Handler:        xssProtector,
	}

	setting = settingManager.Get()
	if len(setting.CertFilePath) != 0 && len(setting.KeyFilePath) != 0 {
		err = server.ListenAndServeTLS(setting.CertFilePath, setting.KeyFilePath)
	}else {
		err = server.ListenAndServe()
	}

	if err != nil {
		HttpLog.Fatal(err)
	}
}
