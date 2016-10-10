package main

import "errors"
import "github.com/naoina/genmai"
import "github.com/cs3238-tsuzu/popcon/file_manager"
import "strconv"
import "os"
import "io/ioutil"

const ContestDir = "./contests/"

type ContestType int

const (
	ContestJOI ContestType = 0

//    ContestICPC ContestType = 1
//    ContestAtCoder ContestType = 2
//    ContestPCK ContestType = 3
)

var ContestTypeToString = map[ContestType]string{
	ContestJOI: "JOI",
}

type Contest struct {
	Cid        int64  `db:"pk" default:""`
	Name       string `default:""`
	StartTime  int64  `default:""`
	FinishTime int64  `default:""`
	Admin      int64  `default:""`
	Type       int64  `default:"0"`
}

func (c *Contest) ProblemAdd(pidx int64, name string, time, mem int64, jtype JudgeType) (*ContestProblem, error) {
	pb, err := mainDB.ContestProblemAdd(c.Cid, pidx, name, time, mem, jtype)

	if err != nil {
		return nil, err
	}

	return mainDB.ContestProblemFind(pb)
}

func (dm *DatabaseManager) CreateContestTable() error {
	err := dm.db.CreateTableIfNotExists(&Contest{})

	if err != nil {
		return err
	}

	dm.db.CreateUniqueIndex(&Contest{}, "name")
	dm.db.CreateIndex(&Contest{}, "start_time")
	dm.db.CreateIndex(&Contest{}, "finish_time")

	return nil
}

func (dm *DatabaseManager) ContestAdd(name string, start int64, finish int64, admin int64, ctype ContestType) (int64, error) {
	res, err := dm.db.DB().Exec("insert into contest (name, start_time, finish_time, admin, type) values (?, ?, ?, ?, ?)", name, start, finish, admin, int64(ctype))

	if err != nil {
		return 0, err
	}

	id, _ := res.LastInsertId()

	fm, err := FileManager.OpenFile(ContestDir+strconv.FormatInt(id, 10), os.O_CREATE|os.O_WRONLY, true)

	if err != nil {
		dm.ContestDelete(id)

		return 0, err
	}

	fm.Close()

	return id, nil
}

func (dm *DatabaseManager) ContestUpdate(cid int64, name string, start int64, finish int64, admin int64, ctype ContestType) error {
	cont := Contest{
		Cid:        cid,
		Name:       name,
		StartTime:  start,
		FinishTime: finish,
		Admin:      admin,
		Type:       int64(ctype),
	}

	_, err := dm.db.Update(&cont)

	if err != nil {
		return err
	}

	return nil
}

func (dm *DatabaseManager) ContestDelete(cid int64) error {
	_, err := dm.db.Delete(&Contest{Cid: cid})

	if err != nil {
		return err
	}

	fm, err := FileManager.OpenFile(ContestDir+strconv.FormatInt(cid, 10), os.O_WRONLY, true)

	if err != nil {
		return err
	}

	defer fm.Close()

	return os.Remove(ContestDir + strconv.FormatInt(cid, 10))
}

func (dm *DatabaseManager) ContestFind(cid int64) (*Contest, error) {
	var resulsts []Contest

	err := dm.db.Select(&resulsts, dm.db.Where("cid", "=", cid))

	if err != nil {
		return nil, err
	}

	if len(resulsts) == 0 {
		return nil, errors.New("Unknown contest")
	}

	return &resulsts[0], nil
}

func (dm *DatabaseManager) ContestDescriptionUpdate(cid int64, desc string) error {
	fm, err := FileManager.OpenFile(ContestDir+strconv.FormatInt(cid, 10), os.O_WRONLY|os.O_TRUNC, true)

	if err != nil {
		return err
	}

	defer fm.Close()

	_, err = fm.Write([]byte(desc))

	return err
}

func (dm *DatabaseManager) ContestDescriptionLoad(cid int64) (string, error) {
	fm, err := FileManager.OpenFile(ContestDir+strconv.FormatInt(cid, 10), os.O_RDONLY, false)

	if err != nil {
		return "", err
	}

	defer fm.Close()

	b, err := ioutil.ReadAll(fm)

	if err != nil {
		return "", err
	}

	return string(b), err
}

func (dm *DatabaseManager) ContestCount(options ...*genmai.Condition) (int64, error) {
	var count int64

	opt := make([]interface{}, len(options)+2)

	opt[0] = dm.db.Count("cid")
	opt[1] = dm.db.From(&Contest{})

	for i := range options {
		opt[i+2] = options[i]
	}

	err := dm.db.Select(&count, opt...)

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (dm *DatabaseManager) ContestList(options ...*genmai.Condition) (*[]Contest, error) {
	var resulsts []Contest

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
