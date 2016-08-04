package main

type ContestParticipation struct {
	Iid     int64  `default:""`
	Cid     int64  `default:""`
	Score   int64  `default:"0"`
	Time    int64  `default:"0"`
	Details string `default:"" size:"5000"`
}

func (dm *DatabaseManager) CreateContestParticipationTable() error {
	err := dm.db.CreateTableIfNotExists(&ContestParticipation{})

	if err != nil {
		return err
	}

	dm.db.CreateUniqueIndex(&ContestParticipation{}, "iid", "cid")
	dm.db.CreateIndex(&ContestParticipation{}, "cid")
	dm.db.CreateIndex(&ContestParticipation{}, "iid")
	dm.db.CreateIndex(&ContestParticipation{}, "score")
	dm.db.CreateIndex(&ContestParticipation{}, "time")

	return nil
}

func (dm *DatabaseManager) ContestParticipationAdd(iid, cid int64) error {
	_, err := dm.db.Insert(&ContestParticipation{Iid: iid, Cid: cid})

	return err
}

func (dm *DatabaseManager) ContestParticipationCheck(iid, cid int64) (bool, error) {
	var cnt int64

	err := dm.db.Select(&cnt, dm.db.Count(), dm.db.From(&ContestParticipation{}), dm.db.Where("iid", "=", iid).And("cid", "=", cid))

	if err != nil {
		return false, err
	}

	if cnt == 0 {
		return false, nil
	} else {
		return true, nil
	}
}

func (dm *DatabaseManager) ContestParticipationRemove(iid, cid int64) error {
	_, err := dm.db.Delete(&ContestParticipation{Iid: iid, Cid: cid})

	return err
}

type RankingHighScoreData struct {
    Sid int64
    Score int64
    Time int64
}

func (dm *DatabaseManager) ContestRankingUpdate(iid, cid int64, pid int64, sm Submission) error {
    raw := dm.db.DB()

    tx, err := raw.Begin()

    if err != nil {
        return err
    }

    defer func() {
        if err := recover(); err != nil {
            tx.Rollback()
        }else {
            tx.Commit()
        }
    }()

    rows, err := tx.Query("select details from contest_participation where iid = ? and cid = ?", iid, cid)
    
    if err != nil {
        panic(err)
    }

    rows.Next()

    var details string

    err = rows.Scan(&details)
    rows.Close()

    if err != nil {
        panic(err)
    }

//    var detMap
}