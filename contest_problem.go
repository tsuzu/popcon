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
    fm, err := FileManager.OpenFile(ContestProblemsDir + strconv.FormatInt(cp.Pid, 10) + "/prob", os.O_WRONLY | os.O_CREATE | os.O_TRUNC, true)

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
    Cases []int `json:"cases"`
    Score int `json:"score"`
}

type CheckerInterface struct {
    Lid int64
    Code string
}

func (cp *ContestProblem) UpdateChecker(lid int64, code string) error {
    fm, err := FileManager.OpenFile(ContestProblemsDir + strconv.FormatInt(cp.Pid, 10) + "/checker", os.O_CREATE | os.O_WRONLY | os.O_TRUNC, true)

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

    if len(b) == 0 {
        return 0, "", nil
    }

    var ci CheckerInterface
    err = json.Unmarshal(b, &ci)

    if err != nil {
        return 0, "", err
    }

    return ci.Lid, ci.Code, nil
}

func (cp *ContestProblem) UpdateTestCaseNames(cases []string, scores []ScoreSet) error {
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

    for i := 0; i < len(cases); i++ {
        os.Rename(ContestProblemsDir + strconv.FormatInt(cp.Pid, 10) + "/cases/" + strconv.FormatInt(int64(i), 10) + "_in", "/tmp/popcon_" + strconv.FormatInt(cp.Pid, 10) + "_" + strconv.FormatInt(int64(i), 10) + "_in")
        os.Rename(ContestProblemsDir + strconv.FormatInt(cp.Pid, 10) + "/cases/" + strconv.FormatInt(int64(i), 10) + "_out", "/tmp/popcon_" + strconv.FormatInt(cp.Pid, 10) + "_" + strconv.FormatInt(int64(i), 10) + "_out")
    }

    defer func() {
        for i := 0; i < len(cases); i++ {
            os.Rename("/tmp/popcon_" + strconv.FormatInt(cp.Pid, 10) + "_" + strconv.FormatInt(int64(i), 10) + "_in", ContestProblemsDir + strconv.FormatInt(cp.Pid, 10) + "/cases/" + strconv.FormatInt(int64(i), 10) + "_in")
            os.Rename("/tmp/popcon_" + strconv.FormatInt(cp.Pid, 10) + "_" + strconv.FormatInt(int64(i), 10) + "_out", ContestProblemsDir + strconv.FormatInt(cp.Pid, 10) + "/cases/" + strconv.FormatInt(int64(i), 10) + "_out")
        }
    }()

    err = os.RemoveAll(ContestProblemsDir + strconv.FormatInt(cp.Pid, 10) + "/cases")

    if err != nil {
        return err
    }

    err = os.MkdirAll(ContestProblemsDir + strconv.FormatInt(cp.Pid, 10) + "/cases", os.ModePerm)

    if err != nil {
        return err
    }

    fp, err := os.OpenFile(ContestProblemsDir + strconv.FormatInt(cp.Pid, 10) + "/cases/data", os.O_CREATE | os.O_WRONLY | os.O_TRUNC, 0644)

    if err != nil {
        return err
    }

    defer fp.Close()

    type TestCaseJson struct {
        CaseNames map[string]string `json:"case_names"`
        Scores []ScoreSet `json:"scores"`
    }    

    var tcj TestCaseJson

    tcj.Scores = scores
    tcj.CaseNames = make(map[string]string)

    for i := range cases {
        tcj.CaseNames[strconv.FormatInt(int64(i), 10)] = cases[i]
    }

    b, err := json.Marshal(tcj)

    if err != nil {
        return err
    }

    _, err = fp.Write(b)

    if err != nil {
        return err
    }

    for i := 0; i < len(cases); i++ {
        f, _ := os.OpenFile(ContestProblemsDir + strconv.FormatInt(cp.Pid, 10) + "/cases/" + strconv.FormatInt(int64(i), 10) + "_in", os.O_CREATE | os.O_WRONLY | os.O_TRUNC, 0644)
        
        f.Close()

        f, _ = os.OpenFile(ContestProblemsDir + strconv.FormatInt(cp.Pid, 10) + "/cases/" + strconv.FormatInt(int64(i), 10) + "_out", os.O_CREATE | os.O_WRONLY | os.O_TRUNC, 0644)
        
        f.Close()
    }

    return nil
}

func (cp *ContestProblem) UpdateTestCase(isInput bool, caseID int, str string) error {
    fm, err := FileManager.OpenFile(ContestProblemsDir + strconv.FormatInt(cp.Pid, 10) + "/.cases_lock", os.O_CREATE | os.O_WRONLY | os.O_TRUNC, true)

    if err != nil {
        return err
    }

    defer fm.Close()

    fileTag := "_in"
    if !isInput {
        fileTag = "_out"
    }

    fp, err := os.OpenFile(ContestProblemsDir + strconv.FormatInt(cp.Pid, 10) + "/cases/" + strconv.FormatInt(int64(caseID), 10) + fileTag, os.O_WRONLY | os.O_TRUNC, 0644)

    if err != nil {
        return err
    }

    defer fp.Close()

    _, err = fp.Write([]byte(str))

    return err
}

func (cp *ContestProblem) LoadTestCase(isInput bool, caseID int) (string, error) {
    fm, err := FileManager.OpenFile(ContestProblemsDir + strconv.FormatInt(cp.Pid, 10) + "/.cases_lock", os.O_RDONLY, false)

    if err != nil {
        return "", err
    }

    defer fm.Close()

    fileTag := "_in"
    if !isInput {
        fileTag = "_out"
    }

    fp, err := os.OpenFile(ContestProblemsDir + strconv.FormatInt(cp.Pid, 10) + "/cases/" + strconv.FormatInt(int64(caseID), 10) + fileTag, os.O_RDONLY, 0644)

    if err != nil {
        return "", err
    }

    defer fp.Close()

    b, err := ioutil.ReadAll(fp)

    if err != nil {
        return "", nil
    }

    return string(b), err
}

func (cp *ContestProblem) LoadTestCases() (*[]TestCase, *[]ScoreSet, error) {
    var scores []ScoreSet
    var cases []TestCase
    
    fm, err := FileManager.OpenFile(ContestProblemsDir + strconv.FormatInt(cp.Pid, 10) + "/.cases_lock", os.O_RDONLY, false)

    if err != nil {
        return nil, nil, err
    }

    defer fm.Close()

    fp, err := os.OpenFile(ContestProblemsDir + strconv.FormatInt(cp.Pid, 10) + "/cases/data", os.O_RDONLY, 0644)

    if err != nil {
        return &cases, &scores, nil
    }

    type TestCaseJson struct {
        CaseNames map[string]string `json:"case_names"`
        Scores []ScoreSet `json:"scores"`
    }

    b, err := ioutil.ReadAll(fp)

    fp.Close()

    if err != nil {
        return nil, nil, err
    }

    var tcj TestCaseJson

    err = json.Unmarshal(b, &tcj)

    if err != nil {
        return &cases, &scores, nil
    }

    scores = tcj.Scores
    
    cases = make([]TestCase, len(tcj.CaseNames))

    for x := range tcj.CaseNames {
        i, err := strconv.ParseInt(x, 10, 32)

        if err != nil {
            return nil, nil, err
        }

        var tcase TestCase
        fp, err := os.OpenFile(ContestProblemsDir + strconv.FormatInt(cp.Pid, 10) + "/cases/" + strconv.FormatInt(i, 10) + "_in", os.O_RDONLY, 0644)

        if err != nil {
            return nil, nil, err
        }

        b, err := ioutil.ReadAll(fp)

        if err != nil {
            return nil, nil, err
        }

        tcase.Input = string(b)

        fp.Close()

        fp, err = os.OpenFile(ContestProblemsDir + strconv.FormatInt(cp.Pid, 10) + "/cases/" + strconv.FormatInt(i, 10) + "_out", os.O_RDONLY, 0644)

        if err != nil {
            return nil, nil, err
        }

        b, err = ioutil.ReadAll(fp)


        if err != nil {
            return nil, nil, err
        }

        tcase.Output = string(b)

        tcase.Name = tcj.CaseNames[x]

        fp.Close()

        cases[i] = tcase
    }

    return &cases, &scores, nil
}

func (cp *ContestProblem) LoadTestCaseInfo(caseId int) (int64, int64, error) {
    fm, err := FileManager.OpenFile(ContestProblemsDir + strconv.FormatInt(cp.Pid, 10) + "/.cases_lock", os.O_RDONLY, false)

    if err != nil {
        return 0, 0, err
    }

    defer fm.Close()

    fp, err := os.OpenFile(ContestProblemsDir + strconv.FormatInt(cp.Pid, 10) + "/cases/" + strconv.FormatInt(int64(caseId), 10) + "_in", os.O_RDONLY, 0644)

    if err != nil {
        return 0, 0, err
    }

    defer fp.Close()

    fi, err := fp.Stat()

    if err != nil {
        return 0, 0, err
    }

    in := fi.Size()

    fp.Close()

    fp, err = os.OpenFile(ContestProblemsDir + strconv.FormatInt(cp.Pid, 10) + "/cases/" + strconv.FormatInt(int64(caseId), 10) + "_out", os.O_RDONLY, 0644)

    if err != nil {
        return 0, 0, err
    }

    fi, err = fp.Stat()

    if err != nil {
        return 0, 0, err
    }

    out := fi.Size()

    return in, out , nil
}

func (cp *ContestProblem) LoadTestCaseNames() (*[]string, *[]ScoreSet, error) {
    var scores []ScoreSet
    var cases []string
    
    fm, err := FileManager.OpenFile(ContestProblemsDir + strconv.FormatInt(cp.Pid, 10) + "/.cases_lock", os.O_RDONLY, false)

    if err != nil {
        return nil, nil, err
    }

    defer fm.Close()

    fp, err := os.OpenFile(ContestProblemsDir + strconv.FormatInt(cp.Pid, 10) + "/cases/data", os.O_RDONLY, 0644)

    if err != nil {
        return &cases, &scores, nil
    }

    type TestCaseJson struct {
        CaseNames map[string]string `json:"case_names"`
        Scores []ScoreSet `json:"scores"`
    }

    b, err := ioutil.ReadAll(fp)

    fp.Close()

    if err != nil {
        return nil, nil, err
    }

    var tcj TestCaseJson

    err = json.Unmarshal(b, &tcj)

    if err != nil {
        return &cases, &scores, nil
    }

    scores = tcj.Scores
    
    cases = make([]string, len(tcj.CaseNames))

    for x := range tcj.CaseNames {
        i, err := strconv.ParseInt(x, 10, 32)

        if err != nil {
            return nil, nil, err
        }

        cases[i] = tcj.CaseNames[x]
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
        dm.ContestProblemDelete(i)

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

    // COUNT(*)が重い
    err := dm.db.Select(&count, dm.db.Count("pid"), dm.db.From(&ContestProblem{}), dm.db.Where("cid", "=", cid))

    if err != nil {
        return 0, err
    }
    
    return count , nil
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

func (dm *DatabaseManager) ContestProblemRemoveAll(cid int64) error {
    _, err := dm.db.DB().Exec("delete from contest_problem where cid = ?", cid)

    return err
}