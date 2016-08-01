package main

import "log"
import "os"
import "fmt"
import "io"

var HttpLog, DBLog *log.Logger
var LO *LogOut

type LogOut struct {
	fp *os.File
}

func (lo LogOut) Write(p []byte) (n int, err error) {
	fmt.Print(string(p))

	if lo.fp != nil {
        lo.fp.Write(p)
    }

	return len(p), nil
}

func CreateLogOut() (*LogOut, error) {
    var fp *os.File
    var err error

	//fp, err = os.OpenFile("./log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)

	return &LogOut{fp}, err
}

func CreateLogger(writer io.Writer) {
    HttpLog = log.New(writer, "popcon: ", log.LstdFlags | log.Llongfile)
    DBLog =  log.New(writer, "mysql: ", log.LstdFlags | log.Llongfile)
}
