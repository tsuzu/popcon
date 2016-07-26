package main

import (
	"time"
	"os"
	"strconv"
	"errors"
	"io/ioutil"
	"github.com/cs3238-tsuzu/popcon/file_manager"
	"encoding/json"
)

// Database manager for Contest and Onlinejudge

//import "errors"

const SubmissionDir = "./submissions/"

type SubmissionStatus int64

const (
    InQueue SubmissionStatus = 0
    Judging SubmissionStatus = 1
    Accepted SubmissionStatus = 2
    WrongAnswer SubmissionStatus = 3
    TimeLimitExceeded SubmissionStatus = 4
    MemoryLimitExceeded SubmissionStatus = 5
    RuntimeError SubmissionStatus = 7
    CompileError SubmissionStatus = 8
    InternalError SubmissionStatus = 9
)

var SubmissionStatusToString = map[SubmissionStatus]string{
    InQueue: "WJ",
    Judging: "JG",
    Accepted: "AC",
    WrongAnswer: "WA",
    TimeLimitExceeded: "TLE",
    MemoryLimitExceeded: "MLE",
    RuntimeError: "RE",
    CompileError: "CE",
    InternalError: "IE",
}

type Submission struct {
    Sid int64 `db:"pk" default:""`
    Pid int64 `default:""` //index
    Cid int64 `default:""` //index, onlinejudge->0
    Iid int64 `default:""` //index
    Lang int64 `default:""`
    Time int64 `default:""` //ms
    Mem int64 `default:""` //KB
    Score int64 `default:""`
    SubmitTime int64 `default:""` //提出日時
    Status int64 `default:""` //index
    Prog uint64 `default:""` //テストケースの進捗状況(完了数<<32 & 全体数)
}

func (dm *DatabaseManager) CreateSubmissionTable() error {
    err := dm.db.CreateTableIfNotExists(&Submission{})

    if err != nil {
        return err
    }

    dm.db.CreateIndex(&Submission{}, "pid")
    dm.db.CreateIndex(&Submission{}, "cid")
    dm.db.CreateIndex(&Submission{}, "iid")
    dm.db.CreateIndex(&Submission{}, "status")

    return nil
}

func (dm *DatabaseManager) SubmissionNew(pid, cid, iid, lang int64, code string) (int64, error) {
    sm := Submission {
        Pid: pid,
        Cid: cid,
        Iid: iid,
        Lang: lang,
        Time: 0,
        Mem: 0,
        Score: 0,
        SubmitTime: time.Now().Unix(),
        Status: int64(InQueue),
        Prog: 0,
    }

    _, err := dm.db.Insert(&sm)
    id := sm.Sid

    if err != nil {
        return 0, err
    }

    err = os.MkdirAll(SubmissionDir + strconv.FormatInt(id, 10), os.ModePerm)

    if err != nil {
        dm.SubmissionRemove(id)

        return 0, err
    }

    fp, err := os.OpenFile(SubmissionDir + strconv.FormatInt(id, 10) + "/msg", os.O_WRONLY | os.O_CREATE, 0644)

    if err != nil {
        dm.SubmissionRemove(id)

        return 0, err
    }

    fp.Close()

    fp, err = os.OpenFile(SubmissionDir + strconv.FormatInt(id, 10) + "/case", os.O_WRONLY | os.O_CREATE, 0644)

    if err != nil {
        dm.SubmissionRemove(id)

        return 0, err
    }

    fp.Close()

    fp, err = os.OpenFile(SubmissionDir + strconv.FormatInt(id, 10) + "/code", os.O_WRONLY | os.O_CREATE, 0644)

    if err != nil {
        dm.SubmissionRemove(id)

        return 0, err
    }

    defer fp.Close()

    _, err = fp.Write([]byte(code))

    if err != nil {
        return 0, err
    }

    return id, nil
}

func (dm *DatabaseManager) SubmissionRemove(sid int64) error {
    sm := Submission{Sid: sid}
    _, err := dm.db.Delete(&sm)
    
    if err != nil {
        return err
    }

    fm1, _ := FileManager.OpenFile(SubmissionDir + strconv.FormatInt(sid, 10) + "/case", os.O_RDONLY, true)
    fm2, _ := FileManager.OpenFile(SubmissionDir + strconv.FormatInt(sid, 10) + "/msg", os.O_RDONLY, true)
    
    defer func () {
        fm1.Close()
        fm2.Close()
    }()

    return os.RemoveAll(SubmissionDir + strconv.FormatInt(sid, 10))
}

func (dm *DatabaseManager) SubmissionFind(sid int64) (*Submission, error) {
    var results []Submission

    err := dm.db.Select(&results, dm.db.Where("sid", "=", sid))

    if err != nil {
        return nil, err
    }

    if len(results) == 0 {
        return nil, errors.New("Unknown submission")
    }

    return &results[0], nil
}

func (dm *DatabaseManager) SubmissionUpdate(sid, time, mem int64, status SubmissionStatus, fin, all int) error {
    sm, err := dm.SubmissionFind(sid)

    if err != nil {
        return err
    }

    sm.Time = time
    sm.Mem = mem
    sm.Status = int64(status)
    sm.Prog = uint64(fin) << 32 & uint64(all)

    _, err = dm.db.Update(&sm)

    return err
}

func (dm *DatabaseManager) SubmissionGetCode(sid int64) (*string, error) {
    fp, err := os.OpenFile(SubmissionDir + strconv.FormatInt(sid, 10) + "/code", os.O_RDONLY, 0644)

    if err != nil {
        return nil, err
    }

    defer fp.Close()

    b, err := ioutil.ReadAll(fp)

    if err != nil {
        return nil, err
    }

    str := string(b)

    return &str, nil
}

func (dm *DatabaseManager) SubmissionGetMsg(sid int64) string {
    fm, err := FileManager.OpenFile(SubmissionDir + strconv.FormatInt(sid, 10) + "/msg", os.O_RDONLY, false)
    
    if err != nil {
        return ""
    }

    defer fm.Close()

    b, err := ioutil.ReadAll(fm)

    if err != nil {
        return ""
    }

    return string(b)
}

func (dm *DatabaseManager) SubmissionSetMsg(sid int64, msg string) error {
    fm, err := FileManager.OpenFile(SubmissionDir + strconv.FormatInt(sid, 10) + "/msg", os.O_WRONLY | os.O_CREATE, true)

    if err != nil {
        return err
    }

    defer fm.Close()
    
    _, err = fm.Write([]byte(msg))

    return err
}

func (dm *DatabaseManager) SubmissionGetCase(sid int64) (*map[int64]SubmissionStatus, error) {
    fm, err := FileManager.OpenFile(SubmissionDir + strconv.FormatInt(sid, 10) + "/case", os.O_RDONLY, false)
    
    if err != nil {
        return nil, err
    }
    
    defer fm.Close()

    b, err := ioutil.ReadAll(fm)

    if err != nil {
        return nil, err
    }

    var ss map[int64]SubmissionStatus
    err = json.Unmarshal(b, &ss)

    if err != nil {
        return nil, err
    }

    return &ss, nil
}

func (dm *DatabaseManager) SubmissionSetCase(sid int64, ss map[int64]SubmissionStatus) error {
    fm, err := FileManager.OpenFile(SubmissionDir + strconv.FormatInt(sid, 10) + "/case", os.O_RDONLY, false)
    
    if err != nil {
        return err
    }
    
    defer fm.Close()

    b, err := json.Marshal(ss)

    if err != nil {
        return err
    }

    _, err = fm.Write(b)

    return err
}

func (dm *DatabaseManager) SubmissionCount(cid, iid, lid int64) {
    //
}