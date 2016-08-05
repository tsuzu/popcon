package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"strconv"
	"time"
	"strings"
	"github.com/cs3238-tsuzu/popcon/file_manager"
	"github.com/naoina/genmai"
)

// Database manager for Contest and Onlinejudge

//import "errors"

const SubmissionDir = "./submissions/"

type SubmissionStatus int64

const (
	InQueue             SubmissionStatus = 0
	Judging             SubmissionStatus = 1
	Accepted            SubmissionStatus = 2
	WrongAnswer         SubmissionStatus = 3
	TimeLimitExceeded   SubmissionStatus = 4
	MemoryLimitExceeded SubmissionStatus = 5
	RuntimeError        SubmissionStatus = 6
	CompileError        SubmissionStatus = 7
	InternalError       SubmissionStatus = 8
)

var SubmissionStatusToString = map[SubmissionStatus]string{
	InQueue:             "WJ",
	Judging:             "JG",
	Accepted:            "AC",
	WrongAnswer:         "WA",
	TimeLimitExceeded:   "TLE",
	MemoryLimitExceeded: "MLE",
	RuntimeError:        "RE",
	CompileError:        "CE",
	InternalError:       "IE",
}

type Submission struct {
	Sid        int64  `db:"pk" default:""`
	Pid        int64  `default:""` //index
	Iid        int64  `default:""` //index
	Lang       int64  `default:""`
	Time       int64  `default:""` //ms
	Mem        int64  `default:""` //KB
	Score      int64  `default:""`
	SubmitTime int64  `default:""` //提出日時
	Status     int64  `default:""` //index
	Prog       uint64 `default:""` //テストケースの進捗状況(完了数<<32 & 全体数)
}

func (dm *DatabaseManager) CreateSubmissionTable() error {
	err := dm.db.CreateTableIfNotExists(&Submission{})

	if err != nil {
		return err
	}

	dm.db.CreateIndex(&Submission{}, "pid")
	dm.db.CreateIndex(&Submission{}, "iid")
	dm.db.CreateIndex(&Submission{}, "status")

	return nil
}

func (dm *DatabaseManager) SubmissionNew(pid, iid, lang int64, code string) (int64, error) {
	sm := Submission{
		Pid:        pid,
		Iid:        iid,
		Lang:       lang,
		Time:       0,
		Mem:        0,
		Score:      0,
		SubmitTime: time.Now().Unix(),
		Status:     int64(InQueue),
		Prog:       0,
	}

	_, err := dm.db.Insert(&sm)
	id := sm.Sid

	if err != nil {
		return 0, err
	}

	err = os.MkdirAll(SubmissionDir+strconv.FormatInt(id, 10), os.ModePerm)

	if err != nil {
		dm.SubmissionRemove(id)

		return 0, err
	}

	fp, err := os.OpenFile(SubmissionDir+strconv.FormatInt(id, 10)+"/msg", os.O_WRONLY|os.O_CREATE, 0644)

	if err != nil {
		dm.SubmissionRemove(id)

		return 0, err
	}

	fp.Close()

	fp, err = os.OpenFile(SubmissionDir+strconv.FormatInt(id, 10)+"/.cases_lock", os.O_WRONLY|os.O_CREATE, 0644)

	if err != nil {
		dm.SubmissionRemove(id)

		return 0, err
	}

	fp.Close()

	err = os.MkdirAll(SubmissionDir+strconv.FormatInt(id, 10)+"/cases", os.ModePerm)

	if err != nil {
		dm.SubmissionRemove(id)

		return 0, err
	}

	err = os.MkdirAll(SubmissionDir+strconv.FormatInt(id, 10), os.ModePerm)

	if err != nil {
		dm.SubmissionRemove(id)

		return 0, err
	}

	fp, err = os.OpenFile(SubmissionDir+strconv.FormatInt(id, 10)+"/code", os.O_WRONLY|os.O_CREATE, 0644)

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

	fm1, _ := FileManager.OpenFile(SubmissionDir+strconv.FormatInt(sid, 10)+"/.cases_lock", os.O_RDONLY, true)
	fm2, _ := FileManager.OpenFile(SubmissionDir+strconv.FormatInt(sid, 10)+"/msg", os.O_RDONLY, true)

	defer func() {
		fm1.Close()
		fm2.Close()
	}()

	return os.RemoveAll(SubmissionDir + strconv.FormatInt(sid, 10))
}

func (dm *DatabaseManager) SubmissionRemoveAll(pid int64) error {
	sml, err := dm.SubmissionList(dm.db.Where("pid", "=", pid))

	if err != nil {
		return err
	}

	for i := range *sml {
		dm.SubmissionRemove((*sml)[i].Sid)
	}

	return nil
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

func (dm *DatabaseManager) SubmissionUpdate(sid, time, mem int64, status SubmissionStatus, fin, all int, score int64) error {
	sm, err := dm.SubmissionFind(sid)

	if err != nil {
		return err
	}

	sm.Time = time
	sm.Mem = mem
	sm.Status = int64(status)
	sm.Prog = uint64(fin) << 32 | uint64(all)
	sm.Score = score

	_, err = dm.db.Update(sm)

	return err
}

func (dm *DatabaseManager) SubmissionGetCode(sid int64) (*string, error) {
	fp, err := os.OpenFile(SubmissionDir+strconv.FormatInt(sid, 10)+"/code", os.O_RDONLY, 0644)

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

func (dm *DatabaseManager) SubmissionGetMsg(sid int64) *string {
	var res string
	fm, err := FileManager.OpenFile(SubmissionDir+strconv.FormatInt(sid, 10)+"/msg", os.O_RDONLY, false)

	if err != nil {
		return &res
	}

	defer fm.Close()

	b, err := ioutil.ReadAll(fm)

	if err != nil {
		return &res
	}

	res = string(b)

	return &res
}

func (dm *DatabaseManager) SubmissionSetMsg(sid int64, msg string) error {
	fm, err := FileManager.OpenFile(SubmissionDir+strconv.FormatInt(sid, 10)+"/msg", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, true)

	if err != nil {
		return err
	}

	defer fm.Close()

	_, err = fm.Write([]byte(msg))

	return err
}

type SubmissionTestCase struct {
	Status SubmissionStatus
	Name string
	Time int64
	Mem int64
}

func (dm *DatabaseManager) SubmissionGetCase(sid int64) (*map[int]SubmissionTestCase, error) {
	res := make(map[int]SubmissionTestCase)

	fm, err := FileManager.OpenFile(SubmissionDir+strconv.FormatInt(sid, 10)+"/.cases_lock", os.O_RDONLY, false)

	if err != nil {
		return nil, err
	}

	defer fm.Close()

	info, err := ioutil.ReadDir(SubmissionDir+strconv.FormatInt(sid, 10)+"/cases")

	if err != nil {
		return nil, err
	}

	for i := range info {
		if !info[i].IsDir() {
			id, err := strconv.ParseInt(info[i].Name(), 10, 64)

			if err != nil {
				continue
			}
			
			fp, err := os.Open(SubmissionDir+strconv.FormatInt(sid, 10) + "/cases/" + info[i].Name())

			if err != nil {
				continue
			}

			defer fp.Close()

			dec := json.NewDecoder(fp)

			var stc SubmissionTestCase

			err = dec.Decode(&stc)

			if err != nil {
				continue
			}

			res[int(id)] = stc
		}
	}

	return &res, nil
}

func (dm *DatabaseManager) SubmissionSetCase(sid int64, caseId int, stc SubmissionTestCase) error {
	fm, err := FileManager.OpenFile(SubmissionDir+strconv.FormatInt(sid, 10)+"/.cases_lock", os.O_WRONLY | os.O_CREATE | os.O_TRUNC, true)

	if err != nil {
		return err
	}

	defer fm.Close()

	fp, err := os.Create(SubmissionDir+strconv.FormatInt(sid, 10)+"/cases/"+strconv.FormatInt(int64(caseId), 10))

	if err != nil {
		return err
	}

	enc := json.NewEncoder(fp)

	return enc.Encode(stc)
}

func (dm *DatabaseManager) SubmissionClearCase(sid int64) error {
	fm, err := FileManager.OpenFile(SubmissionDir+strconv.FormatInt(sid, 10)+"/.cases_lock", os.O_WRONLY | os.O_CREATE | os.O_TRUNC, true)

	if err != nil {
		return err
	}

	defer fm.Close()

	err = os.RemoveAll(SubmissionDir+strconv.FormatInt(sid, 10)+"/cases/")

	if err != nil {
		return err
	}

	return os.MkdirAll(SubmissionDir+strconv.FormatInt(sid, 10)+"/cases/", os.ModePerm)
}

func (dm *DatabaseManager) SubmissionList(options ...*genmai.Condition) (*[]Submission, error) {
    var resulsts []Submission

    opt := make([]interface{}, len(options))

    for i := range options {
        opt[i] = options[i]
    }

    err := dm.db.Select(&resulsts, opt...)

    if err != nil {
        return nil, err
    }

    return &resulsts, nil
}

type SubmissionView struct {
	SubmitTime int64
	Cid int64
	Pidx int64
	Name string
	Uid string
	UserName string
	Lang string
	Score int64
	Status string
	Time int64
	Mem int64
	Sid int64
}

func (dm *DatabaseManager) submissionViewQueryCreate(query string, cid, iid, lid, pidx, stat int64, order string, offset, limit int64) (*string, error) {
	conditions := make([]string, 0, 5)

	if cid != -1 {
		conditions = append(conditions, "contest_problem.cid = " + strconv.FormatInt(cid, 10) + " ")
	}

	if iid != -1 {
		conditions = append(conditions, "user.iid = " + strconv.FormatInt(iid, 10) + " ")
	}

	if pidx != -1 {
        if cid == -1 {
            return nil, errors.New("You must set cid to set pidx")
        }

		conditions = append(conditions, "contest_problem.pidx = " + strconv.FormatInt(pidx, 10) + " ")
	}

	if stat != -1 {
		conditions = append(conditions, "submission.status = " + strconv.FormatInt(stat, 10) + " ")
	}

	where := strings.Join(conditions, "and ")

	if len(where) != 0 {
		where = "where " + where
	}

    var lim string
    if offset != -1 {
        lim = "limit " + strconv.FormatInt(offset, 10)
        if limit != -1 {
            lim = lim + ", " + strconv.FormatInt(limit, 10)
        }
    }else {
        if limit != -1 {
            lim = "limit " + strconv.FormatInt(limit, 10)
        }
    }

    query += where + order + lim

    return &query, nil
}

func (dm *DatabaseManager) SubmissionViewCount(cid, iid, lid, pidx, stat int64) (int64, error) {
    queryBase := "select count(submission.sid) from submission inner join contest_problem on submission.pid = contest_problem.pid inner join user on submission.iid = user.iid inner join language on submission.lang = language.lid "

    query, err := dm.submissionViewQueryCreate(queryBase, cid, iid, lid, pidx, stat, "", -1, -1)

    if err != nil {
        return 0, err
    }

    rows, err := dm.db.DB().Query(*query)

    if err != nil {
        return 0, err
    }

	defer rows.Close()

    rows.Next()

    var cnt int64
    err = rows.Scan(&cnt)

    if err != nil {
        return 0, err
    }


    return cnt, err
}

func (dm *DatabaseManager) SubmissionViewList(cid, iid, lid, pidx, stat, offset, limit int64) (*[]SubmissionView, error) {
	queryBase := "select submission.submit_time, contest_problem.cid, contest_problem.pidx, contest_problem.name, user.uid, user.user_name, language.name, submission.score, submission.status, submission.prog, submission.time, submission.mem, submission.sid from submission inner join contest_problem on submission.pid = contest_problem.pid inner join user on submission.iid = user.iid inner join language on submission.lang = language.lid "

	query, err := dm.submissionViewQueryCreate(queryBase, cid, iid, lid, pidx, stat, "order by submission.sid desc ", offset, limit)
    
    if err != nil {
        return nil, err
    }

    rows, err := dm.db.DB().Query(*query)
	
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	views := make([]SubmissionView, 0, 50)

	for rows.Next() {
		var sv SubmissionView

		var status int64
		var prog uint64
		rows.Scan(&sv.SubmitTime, &sv.Cid, &sv.Pidx, &sv.Name, &sv.Uid, &sv.UserName, &sv.Lang, &sv.Score, &status, &prog, &sv.Time, &sv.Mem, &sv.Sid)

		if status == int64(Judging) {
            all := prog & ((uint64(1) << 32) - 1)
            per := prog >> 32

			sv.Status = strconv.FormatInt(int64(per), 10) + "/" + strconv.FormatInt(int64(all), 10)
		}else {
			sv.Status = SubmissionStatusToString[SubmissionStatus(status)]
		}

        if status != int64(Accepted) && status != int64(WrongAnswer) {
            sv.Mem = -1
            sv.Time = -1
            sv.Score = -1
        }

        views = append(views, sv)
	}

	return &views, nil
}

type SubmissionViewEach struct {
	SubmissionView
	HighlightType string
    Iid int64
}

func (dm *DatabaseManager) SubmissionViewFind(sid int64) (*SubmissionViewEach, error) {
	query := "select submission.submit_time, contest_problem.cid, contest_problem.pidx, contest_problem.name, user.uid, user.user_name, language.name, submission.score, submission.status, submission.prog, submission.time, submission.mem, submission.sid, language.highlight_type, submission.iid from submission inner join contest_problem on submission.pid = contest_problem.pid inner join user on submission.iid = user.iid inner join language on submission.lang = language.lid where submission.sid = " + strconv.FormatInt(sid, 10)

    rows, err := dm.db.DB().Query(query)
	
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	rows.Next()
	
	var sv SubmissionViewEach
	var status int64
	var prog uint64
	
	err = rows.Scan(&sv.SubmitTime, &sv.Cid, &sv.Pidx, &sv.Name, &sv.Uid, &sv.UserName, &sv.Lang, &sv.Score, &status, &prog, &sv.Time, &sv.Mem, &sv.Sid, &sv.HighlightType, &sv.Iid)

	if err != nil {
		return nil, err
	}

	if status == int64(Judging) {
    	all := prog & ((uint64(1) << 32) - 1)
        per := prog >> 32

		sv.Status = strconv.FormatInt(int64(per), 10) + "/" + strconv.FormatInt(int64(all), 10)
	}else {
		sv.Status = SubmissionStatusToString[SubmissionStatus(status)]
	}

    if status != int64(Accepted) && status != int64(WrongAnswer) {
        sv.Mem = -1
        sv.Time = -1
        sv.Score = -1
    }

	return &sv, nil
}

func (dm *DatabaseManager) SubmissionGetPid(sid int64) (int64, error) {
	rows, err := dm.db.DB().Query("select pid from submission where sid = ?", sid)

	if err != nil {
		return 0, err
	}

	defer rows.Close()

	var res int64
	rows.Next()
	err = rows.Scan(&res)

	return res, err
}
