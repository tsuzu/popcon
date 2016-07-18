package main

import "errors"

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
    Cid int64 `db:"pk" default:""`
    Name string `default:""`
    StartTime int64 `default:""`
    FinishTime int64 `default:""`
    Admin int64 `default:""`
    Type ContestType `default:"0"`
}

func (c *Contest) ProblemAdd(pidx int64, name string, time, mem int64) (*ContestProblem, error) {
    pb, err := mainDB.ContestProblemNew(c.Cid, pidx, name, time, mem)

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

func (dm *DatabaseManager) ContestNew(name string, start int64, finish int64, admin int64, ctype ContestType) (int64, error) {
    id, err := dm.db.Insert(Contest{
        Name: name,
        StartTime: start,
        FinishTime: finish,
        Admin: admin,
        Type: ctype,
    })

    if err != nil {
        return 0, err
    }

    return id, nil
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

func (dm *DatabaseManager) ContestCount(options ...interface{}) (int64, error) {
    var count int64

    err := dm.db.Select(&count, dm.db.Count(), dm.db.From(&Contest{}), options)

    if err != nil {
        return 0, err
    }

    return count, nil
}

func (dm *DatabaseManager) ContestList(options ...interface{}) (*[]Contest, error) {
    var resulsts []Contest

    err := dm.db.Select(&resulsts, options)

    if err != nil {
        return nil, err
    }

    return &resulsts, nil
}