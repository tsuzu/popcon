package main

import (
	"errors"
	"time"
)

const ContestProblemsDir = "./contest_problems/"

type ContestProblem struct {
    Pid int64 `db:"pk"`
    Cid int64 `default:""`
    Pidx int64 `default:""`
    Name string `default:"" size:"256"`
    Time int64 `default:""` // ms
    Mem int64 `default:""` // KB
    LastModified int64 `default:""`
}

func (dm *DatabaseManager) CreateContestProblemTable() error {
    err := dm.db.CreateTableIfNotExists(&ContestProblem{})

    if err != nil {
        return err
    }

    dm.db.CreateIndex(&ContestProblem{}, "cid")
    dm.db.CreateIndex(&ContestProblem{}, "pidx")

    return nil
}

func (dm *DatabaseManager) ContestProblemNew(cid, pidx int64, name string, timel, mem int64) (int64, error) {
    prob := ContestProblem{
        Cid: cid,
        Pidx: pidx,
        Name: name,
        Time: timel,
        Mem: mem,
        LastModified: time.Now().Unix(),
    }

    return dm.db.Insert(prob)
}

func (dm *DatabaseManager) ContestProblemDelete(pid int64) error {
    prob := ContestProblem{ Pid: pid }

    _, err := dm.db.Delete(&prob)

    return err
}

func (dm *DatabaseManager) ContestProblemFind(pid int64) (*ContestProblem, error) {
    var resulsts []ContestProblem

    err := dm.db.Select(&resulsts, dm.db.Where("pid", "=", pid))

    if err != nil {
        return nil, err
    }

    if len(resulsts) == 0 {
        return nil, errors.New("Unknown problem")
    }

    return &resulsts[0], nil
}