package main

type ContestParticipation struct {
    Iid int64 `default:""`
    Cid int64 `default:""`
}

func (dm *DatabaseManager) CreateContestParticipationManager() error {
    err := dm.db.CreateTableIfNotExists(&ContestParticipation{})

    if err != nil {
        return err
    }

    dm.db.CreateUniqueIndex(&ContestParticipation{}, "iid", "cid")
    dm.db.CreateIndex(&ContestParticipation{}, "cid")
    dm.db.CreateIndex(&ContestParticipation{}, "iid")
    
    return nil
}

func (dm *DatabaseManager) ContestParticipationAdd(iid, cid int64) error {
    _, err := dm.db.Insert(&ContestParticipation{iid, cid})

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
    }else {
        return true, nil
    }
}

func (dm *DatabaseManager) ContestParticipationRemove(iid, cid int64) error {
    _, err := dm.db.Delete(&ContestParticipation{iid, cid})

    return err
}