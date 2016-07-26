package main

import (
	"errors"
)

type Language struct {
    Lid int64 `db:"pk"`
    Name string `default:""`
    HighlightType string `default:""`
    Active bool `default:"true"`
}

func (dm *DatabaseManager) CreateLanguageTable() error {
    err := dm.db.CreateTableIfNotExists(&Language{})

    if err != nil {
        return err
    }

    dm.db.CreateIndex(&Language{}, "active")

    return nil
}

func (dm *DatabaseManager) LanguageAdd(name, highlightType string) (int64, error) {
    lang := Language{
        Name: name,
        HighlightType: highlightType,
        Active: true,
    }

    _, err := dm.db.Insert(&lang)

    return lang.Lid, err
}

func (dm *DatabaseManager) LanguageUpdate(lid int64, name, highlightType string, active bool) error {
    _, err := dm.db.Update(&Language{
        Lid: lid,
        Name: name,
        HighlightType: highlightType,
        Active: active,
    })

    return err
}

func (dm *DatabaseManager) LanguageFind(lid int64) (*Language, error) {
    var resulsts []Language

    err := dm.db.Select(&resulsts, dm.db.Where("lid", "=", lid))

    if err != nil {
        return nil, err
    }

    if len(resulsts) == 0 {
        return nil, errors.New("Unknown language")
    }

    return &resulsts[0], nil
}

func (dm *DatabaseManager) LanguageList() (*[]Language, error) {
    var resulsts []Language

    err := dm.db.Select(&resulsts)

    if err != nil {
        return nil, err
    }

    return &resulsts, nil
}