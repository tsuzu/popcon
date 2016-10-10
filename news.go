package main

import "errors"
import "time"
import "github.com/naoina/genmai"

// "create table if not exists news (text varchar(256), unixTime int, index uti(unixTime))"
// News contains news showed on "/"
type News struct {
	Text     string `default:""`
	UnixTime int64  `default:"" size:"1024"`
}

func (dm *DatabaseManager) CreateNewsTable() error {
	err := dm.db.CreateTableIfNotExists(&News{})

	if err != nil {
		return err
	}

	dm.db.CreateIndex(&News{}, "unixTime")

	return nil
}

// AddNews adds a news displayed on "/"
// len(text) <= 256
func (dm *DatabaseManager) NewsAdd(text string) error {
	if len(text) > 256 {
		return errors.New("len(text) > 256")
	}

	_, err := dm.db.Insert(&News{text, time.Now().Unix()})

	//_, err := dm.db.Exec("insert into news (text, unixTime) values(?, unix_timestamp(now()))", text)

	return err
}

// AddNewsWithTime adds a news displayed on "/" with unixtime
// len(text) <= 256
func (dm *DatabaseManager) NewsAddWithTime(text string, unixTime time.Time) error {
	if len(text) > 256 {
		return errors.New("len(text) > 256")
	}

	_, err := dm.db.Insert(&News{text, unixTime.Unix()})

	return err
}

// GetNews returns `showedNewCount` recent news of
func (dm *DatabaseManager) NewsGet() ([]News, error) {
	var resulsts []News
	err := dm.db.Select(&resulsts, dm.db.OrderBy("unix_time", genmai.DESC).Limit(dm.showedNewCount))

	if err != nil {
		return nil, err
	}

	return resulsts, nil
}
