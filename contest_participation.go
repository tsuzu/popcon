package main

import (
	"encoding/json"
	"strconv"
)

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
	Sid   int64
	Score int64
	Time  int64
}

func (dm *DatabaseManager) ContestRankingCount(cid int64) (int64, error) {
    var cnt int64
    err := dm.db.Select(&cnt, dm.db.Count("cid", "iid"), dm.db.From("contest_participation"), dm.db.Where("cid", "=", cid))

    return cnt, err
}

type RankingRow struct {
    Uid string
    UserName string
    Score int64
    Time int64
    Probs map[int64]RankingHighScoreData
}

func (dm *DatabaseManager) ContestRankingList(cid int64, offset int64, limit int64) (*[]RankingRow, error) {
    rows, err := dm.db.DB().Query("select user.uid, user.user_name, contest_participation.score, contest_participation.time, contest_participation.details from contest_participation inner join user on user.iid = contest_participation.iid where cid = ? order by contest_participation.score desc, contest_participation.time limit ?, ?", cid, offset,limit)

    if err != nil {
        return nil, err
    }

    defer rows.Close()
    

	mapStrToInt := func(arg map[string]RankingHighScoreData) map[int64]RankingHighScoreData {
		m := make(map[int64]RankingHighScoreData)

		for k, v := range arg {
			idx, err := strconv.ParseInt(k, 10, 64)

			if err == nil {
				m[idx] = v
			}
		}

		return m
    }

    resulsts := make([]RankingRow, 0, 50)
    for rows.Next() {
        var rr RankingRow
        var str string
        
        err := rows.Scan(&rr.Uid, &rr.UserName, &rr.Score, &rr.Time, &str)

        if err != nil {
            return nil, err
        }

        var detStrMap map[string]RankingHighScoreData

        err = json.Unmarshal([]byte(str), &detStrMap)

        if err != nil {
            return nil, err
        }

        rr.Probs = mapStrToInt(detStrMap)

        resulsts = append(resulsts, rr)
    }

    return &resulsts, err
}

func (dm *DatabaseManager) ContestRankingUpdate(sm Submission) (rete error) {
	raw := dm.db.DB()

	tx, err := raw.Begin()

	if err != nil {
		return err
	}

	cp, err := dm.ContestProblemFind(sm.Pid)

	if err != nil {
		return err
	}

	defer func() {
		if err := recover(); err != nil {
			rete = err.(error)
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	rows, err := tx.Query("select score, time, details from contest_participation where iid = ? and cid = ?", sm.Iid, cp.Cid)

	if err != nil {
		panic(err)
	}

	rows.Next()

	var totalScore, totalTime int64
	var details string

	err = rows.Scan(&totalScore, &totalTime, &details)
	rows.Close()

	if err != nil {
		panic(err)
	}

	var detStrMap map[string]RankingHighScoreData

	json.Unmarshal([]byte(details), &detStrMap) // if some error happens, detStrMap should be empty.

	detMap := func() map[int64]RankingHighScoreData {
		m := make(map[int64]RankingHighScoreData)

		for k, v := range detStrMap {
			idx, err := strconv.ParseInt(k, 10, 64)

			if err == nil {
				m[idx] = v
			}
		}

		return m
	}()

	val, has := detMap[sm.Pid]

	if !has {
		detMap[sm.Pid] = RankingHighScoreData{sm.Sid, sm.Score, sm.Time}

		totalScore += sm.Score

		if totalTime < sm.Time {
			totalTime = sm.Time
		}
	} else {
		if val.Sid == sm.Sid {
			rows, err := raw.Query("select sid, score, time from submission where pid = ? and cid = ? order by score desc, time limit 1", sm.Pid, cp.Cid)

			if err != nil {
				panic(err)
			}

			rows.Next()

			var sid, score, time int64
			err = rows.Scan(&sid, &score, &time)

            rows.Close()

			if err != nil {
                detMap[sm.Pid] = RankingHighScoreData{0, 0, 0}
			} else {
                detMap[sm.Pid] = RankingHighScoreData{sid, score, time}
			}
			
            var scoreSum, maxTime int64
			for i := range detMap {
				scoreSum += detMap[i].Score

				if maxTime < detMap[i].Time {
					maxTime = detMap[i].Time
				}
			}
            totalScore = scoreSum
            totalTime = maxTime
    	} else {
			if val.Score < sm.Score {
				detMap[sm.Pid] = RankingHighScoreData{sm.Sid, sm.Score, sm.Time}

				var scoreSum, maxTime int64
				for i := range detMap {
					scoreSum += detMap[i].Score

					if maxTime < detMap[i].Time {
						maxTime = detMap[i].Time
					}
				}
                totalScore = scoreSum
                totalTime = maxTime
			} else if val.Sid > sm.Sid {
				if val.Score == sm.Score {
					detMap[sm.Pid] = RankingHighScoreData{sm.Sid, sm.Score, sm.Time}

					var maxTime int64
					for i := range detMap {
						if maxTime < detMap[i].Time {
							maxTime = detMap[i].Time
						}
					}
                    totalTime = maxTime
				}
			}
		}
	}

    detStrMap = func() map[string]RankingHighScoreData {
		m := make(map[string]RankingHighScoreData)

		for k, v := range detMap {
            m[strconv.FormatInt(k, 10)] = v
		}

		return m
	}()

    b, err := json.Marshal(detStrMap)

    if err != nil {
        panic(err)
    }

  	_, err = tx.Exec("update contest_participation set score = ?, time = ?, details = ? where cid = ? and iid = ?", totalScore, totalTime, string(b), sm.Iid, cp.Cid)
    
    if err != nil {
        panic(err)
    }

    return nil
}

func (dm *DatabaseManager) ContestRankingCheckProblem(cid int64) (rete error) {
	raw := dm.db.DB()

	tx, err := raw.Begin()

	if err != nil {
		return err
	}

	cps, err := dm.ContestProblemList(cid)

	if err != nil {
		return err
	}

    probExists := make(map[int64]bool)

    for i := range *cps {
        probExists[(*cps)[i].Pid] = true
    }

	defer func() {
		if err := recover(); err != nil {
			rete = err.(error)
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	rows, err := tx.Query("select iid, score, time, details from contest_participation where cid = ?", cid)

	if err != nil {
		panic(err)
	}

    defer rows.Close()

	var iid, totalScore, totalTime int64
	var details string

	mapStrToInt := func(arg map[string]RankingHighScoreData) map[int64]RankingHighScoreData {
		m := make(map[int64]RankingHighScoreData)

		for k, v := range arg {
			idx, err := strconv.ParseInt(k, 10, 64)

			if err == nil {
				m[idx] = v
			}
		}

		return m
    }

    mapIntToStr := func(arg map[int64]RankingHighScoreData) map[string]RankingHighScoreData {
		m := make(map[string]RankingHighScoreData)

		for k, v := range arg {
            m[strconv.FormatInt(k, 10)] = v
		}

		return m
	}

    for rows.Next() {
        err = rows.Scan(&iid, &totalScore, &totalTime, &details)

        if err != nil {
            panic(err)
        }

        var detStrMap map[string]RankingHighScoreData

	    err = json.Unmarshal([]byte(details), &detStrMap)

    	if err != nil {
	    	panic(err)
    	}

        detMap := mapStrToInt(detStrMap)

        totalScore = 0
        totalTime = 0
        for k, v := range detMap {
            if _, has := probExists[k]; !has {
                delete(detMap, k)
            }
            totalScore += v.Score
            
            if totalTime < v.Time {
                totalTime = v.Time
            }
        }

        detStrMap = mapIntToStr(detMap)

        b, err := json.Marshal(detStrMap)

        if err != nil {
            panic(err)
        }

        _, err = tx.Exec("update contest_participation set score = ?, time = ?, details = ? where cid = ? and iid = ?", totalScore, totalTime, string(b), iid, cid)
    
        if err != nil {
            panic(err)
        }
    }

    return nil
}

