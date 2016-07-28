package main

import (
	"errors"
	"time"
	"strconv"
	"os"
    "github.com/cs3238-tsuzu/popcon/file_manager"
	"io/ioutil"
	"encoding/json"
	"github.com/naoina/genmai"
)

const ContestProblemsDir = "./contest_problems/"

type JudgeType int

const (
    JudgePerfectMatch JudgeType = 0
    JudgeRunningCode JudgeType = 1
)

type ContestProblem struct {
    Pid int64 `db:"pk"`
    Cid int64 `default:""`
    Pidx int64 `default:""`
    Name string `default:"" size:"256"`
    Time int64 `default:""` // Second
    Mem int64 `default:""` // MB
    LastModified int64 `default:""`
    Score int `default:""`
    Type int `default:""`
}

func (cp *ContestProblem) UpdateStatement(text string) error {
    fm, err := FileManager.OpenFile(ContestProblemsDir + strconv.FormatInt(cp.Pid, 10) + "/prob", os.O_WRONLY | os.O_CREATE, true)

    if err != nil {
        return err
    }

    fm.Write([]byte(text))

    fm.Close()

    return nil
}

func (cp *ContestProblem) LoadStatement() (*string, error) {
    fm, err := FileManager.OpenFile(ContestProblemsDir + strconv.FormatInt(cp.Pid, 10) + "/prob", os.O_RDONLY | os.O_CREATE, false)

    if err != nil {
        return nil, err
    }
    defer fm.Close()

    b, err := ioutil.ReadAll(fm)

    if err != nil {
        return nil, err
    }

    str := string(b)

    return &str, nil
}

type TestCase struct {
    Name string
    Input string
    Output string
}

type ScoreSet struct {
    Name string `json:"name"`
    Cases []int `json:"cases"`
    Score int `json:"score"`
}

type CheckerInterface struct {
    Lid int64
    Code string
}

func (cp *ContestProblem) UpdateChecker(lid int64, code string) error {
    fm, err := FileManager.OpenFile(ContestProblemsDir + strconv.FormatInt(cp.Pid, 10) + "/checker", os.O_CREATE | os.O_WRONLY, true)

    if err != nil {
        return err
    }

    defer fm.Close()

    b, err := json.Marshal(CheckerInterface{lid, code})

    if err != nil {
        return err
    }

    _, err = fm.Write(b)

    return err
}

func (cp *ContestProblem) LoadChecker() (int64, string, error) {
    fm, err := FileManager.OpenFile(ContestProblemsDir + strconv.FormatInt(cp.Pid, 10) + "/checker", os.O_RDONLY, false)

    if err != nil {
        return 0, "", err
    }

    defer fm.Close()

    b, err := ioutil.ReadAll(fm)

    if err != nil {
        return 0, "", err
    }

    var ci CheckerInterface
    err = json.Unmarshal(b, &ci)

    if err != nil {
        return 0, "", err
    }

    return ci.Lid, ci.Code, nil
}

func (cp *ContestProblem) UpdateTestCases(cases []TestCase, scores []ScoreSet) error {
    scoreSum := 0
    for i := range scores {
        scoreSum += scores[i].Score
    }

    cp.Score = scoreSum
    err := mainDB.ContestProblemUpdate(*cp)

    if err != nil {
        return err
    }
    
    fm, err := FileManager.OpenFile(ContestProblemsDir + strconv.FormatInt(cp.Pid, 10) + "/.cases_lock", os.O_CREATE | os.O_WRONLY, true)

    if err != nil {
        return err
    }

    defer fm.Close()

    err = os.RemoveAll(ContestProblemsDir + strconv.FormatInt(cp.Pid, 10) + "/cases")

    if err != nil {
        return err
    }

    err = os.MkdirAll(ContestProblemsDir + strconv.FormatInt(cp.Pid, 10) + "/cases", os.ModePerm)

    if err != nil {
        return err
    }

    for i := range cases {
        fp, err := os.OpenFile(ContestProblemsDir + strconv.FormatInt(cp.Pid, 10) + "/cases/" + strconv.FormatInt(int64(i), 10) + "_in", os.O_CREATE | os.O_WRONLY, 0644)

        if err != nil {
            return err
        }

        _, err = fp.Write([]byte(cases[i].Input))

        if err != nil {
            return err
        }

        fp.Close()

        fp, err = os.OpenFile(ContestProblemsDir + strconv.FormatInt(cp.Pid, 10) + "/cases/" + strconv.FormatInt(int64(i), 10) + "_out", os.O_CREATE | os.O_WRONLY, 0644)

        if err != nil {
            return err
        }

        _, err = fp.Write([]byte(cases[i].Output))

        if err != nil {
            return err
        }

        fp.Close()        
    }

    fp, err := os.OpenFile(ContestProblemsDir + strconv.FormatInt(cp.Pid, 10) + "/cases/data", os.O_CREATE | os.O_WRONLY, 0644)

    if err != nil {
        return err
    }

    defer fp.Close()

    type TestCaseJson struct {
        CaseNames map[int]string `json:"case_names"`
        Scores []ScoreSet `json:"scores"`
    }    

    var tcj TestCaseJson

    tcj.Scores = scores
    
    for i := range cases {
        tcj.CaseNames[i] = cases[i].Name
    }

    json.Marshal(tcj)

    return nil
}

func (cp *ContestProblem) LoadTestCases() (*[]TestCase, *[]ScoreSet, error) {
    fm, err := FileManager.OpenFile(ContestProblemsDir + strconv.FormatInt(cp.Pid, 10) + "/.cases_lock", os.O_RDONLY, false)

    if err != nil {
        return nil, nil, err
    }

    defer fm.Close()

    fp, err := os.OpenFile(ContestProblemsDir + strconv.FormatInt(cp.Pid, 10) + "/cases/data", os.O_RDONLY, 0644)

    if err != nil {
        return nil, nil, err
    }

    defer fp.Close()

    type TestCaseJson struct {
        CaseNames map[int]string `json:"case_names"`
        Scores []ScoreSet `json:"scores"`
    }

    b, err := ioutil.ReadAll(fp)

    if err != nil {
        return nil, nil, err
    }

    var tcj TestCaseJson

    json.Unmarshal(b, &tcj)

    scores := tcj.Scores
    
    cases := make([]TestCase, len(tcj.CaseNames))

    for i := range tcj.CaseNames {
        var tcase TestCase
        fp, err := os.OpenFile(ContestProblemsDir + strconv.FormatInt(cp.Pid, 10) + "/cases/" + strconv.FormatInt(int64(i), 10) + "_in", os.O_RDONLY, 0644)

        if err != nil {
            return nil, nil, err
        }

        defer fp.Close()

        b, err := ioutil.ReadAll(fp)

        if err != nil {
            return nil, nil, err
        }

        tcase.Input = string(b)

        fp.Close()

        fp, err = os.OpenFile(ContestProblemsDir + strconv.FormatInt(cp.Pid, 10) + "/cases/" + strconv.FormatInt(int64(i), 10) + "_out", os.O_RDONLY, 0644)

        if err != nil {
            return nil, nil, err
        }

        defer fp.Close()

        b, err = ioutil.ReadAll(fp)


        if err != nil {
            return nil, nil, err
        }

        tcase.Output = string(b)

        tcase.Name = tcj.CaseNames[i]

        fp.Close()

        cases[i] = tcase
    }

    return &cases, &scores, nil
}

func (dm *DatabaseManager) CreateContestProblemTable() error {
    err := dm.db.CreateTableIfNotExists(&ContestProblem{})

    if err != nil {
        return err
    }

    dm.db.CreateIndex(&ContestProblem{}, "cid")
    dm.db.CreateIndex(&ContestProblem{}, "pidx")
    dm.db.CreateUniqueIndex(&ContestProblem{}, "pidx", "cid")

    return nil
}

func (dm *DatabaseManager) ContestProblemNew(cid, pidx int64, name string, timel, mem int64, jtype JudgeType) (int64, error) {
    prob := ContestProblem{
        Cid: cid,
        Pidx: pidx,
        Name: name,
        Time: timel,
        Mem: mem,
        LastModified: time.Now().Unix(),
        Type: int(jtype),
    }

    _, err := dm.db.Insert(&prob)

    i := prob.Pid

    if err != nil {
        return 0, err
    }

    err = os.MkdirAll(ContestProblemsDir + strconv.FormatInt(i, 10) + "/cases/", os.ModePerm)

    if err != nil {
        //dm.ContestProblemDelete(i)

        return 0, err
    }

    fm, err := FileManager.OpenFile(ContestProblemsDir + strconv.FormatInt(i, 10) + "/.cases_lock", os.O_WRONLY | os.O_CREATE, true)

    if err != nil {
        dm.ContestProblemDelete(i)

        return 0, err
    }

    fm.Close()

    fm, err = FileManager.OpenFile(ContestProblemsDir + strconv.FormatInt(i, 10) + "/checker", os.O_WRONLY | os.O_CREATE, true)

    if err != nil {
        dm.ContestProblemDelete(i)

        return 0, err
    }

    fm.Close()

    fm, err = FileManager.OpenFile(ContestProblemsDir + strconv.FormatInt(i, 10) + "/prob", os.O_WRONLY | os.O_CREATE, true)


    if err != nil {
        dm.ContestProblemDelete(i)

        return 0, err
    }

    fm.Close()

    return i, err
}

func (dm *DatabaseManager) ContestProblemUpdate(prob ContestProblem) error {
    _, err := dm.db.Update(&prob)

    return err
}

func (dm *DatabaseManager) ContestProblemDelete(pid int64) error {
    prob := ContestProblem{ Pid: pid }

    _, err := dm.db.Delete(&prob)

    if err != nil {
        return err
    }

    fm1, _ := FileManager.OpenFile(ContestProblemsDir + strconv.FormatInt(pid, 10) + "/prob.txt", os.O_WRONLY | os.O_CREATE, true)
    fm2, _ := FileManager.OpenFile(ContestProblemsDir + strconv.FormatInt(pid, 10) + "/.cases_lock", os.O_WRONLY | os.O_CREATE, true)
    fm3, _ := FileManager.OpenFile(ContestProblemsDir + strconv.FormatInt(pid, 10) + "/checker", os.O_WRONLY | os.O_CREATE, true)

    defer func() {
        if fm1 != nil {
            fm1.Close()
        }
        if fm2 != nil {
            fm2.Close()
        }
        if fm3 != nil {
            fm3.Close()
        }
    }()

    err = os.RemoveAll(ContestProblemsDir + strconv.FormatInt(pid, 10))

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

func (dm *DatabaseManager) ContestProblemFind2(cid, pidx int64) (*ContestProblem, error) {
    var resulsts []ContestProblem

    err := dm.db.Select(&resulsts, dm.db.Where("pidx", "=", pidx).And("cid", "=", cid))

    if err != nil {
        return nil, err
    }

    if len(resulsts) == 0 {
        return nil, errors.New("Unknown problem")
    }

    return &resulsts[0], nil
}

func (dm *DatabaseManager) ContestProblemList(cid int64) (*[]ContestProblem, error) {
    var results []ContestProblem

    err := dm.db.Select(&results, dm.db.Where("cid", "=", cid), dm.db.OrderBy("pidx", genmai.ASC))

    if err != nil {
        return nil, err
    }

    return &results, nil
}

func (dm *DatabaseManager) ContestProblemCount(cid int64) (int64, error) {
    var count int64    

    err := dm.db.Select(&count, dm.db.Count(), dm.db.From(&ContestProblem{}), dm.db.Where("cid", "=", cid))

    if err != nil {
        return 0, err
    }

    return count, nil
}

type ContestProblemLight struct {
    Pidx int64
    Name string
}

func (dm *DatabaseManager) ContestProblemListLight(cid int64) (*[]ContestProblemLight, error) {
    results := make([]ContestProblemLight, 0, 50)

    rows, err := dm.db.DB().Query("select pidx, name from contest_problem where cid = ?", cid)

    if err != nil {
        return nil, err
    }

    for rows.Next() {
        var cpl ContestProblemLight

        rows.Scan(&cpl.Pidx, &cpl.Name)

        results = append(results, cpl)
    }

    return &results, nil
}