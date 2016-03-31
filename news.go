package main

import "errors"
import "time"

// News contains news showed on "/"
type News struct {
    Text string
    UnixTime string
}

// AddNews adds a news displayed on "/"
// len(text) <= 256
func (dm *DatabaseManager) AddNews(text string) error {
    if len(text) > 256 {
        return errors.New("len(text) > 256")
    }
    
    _, err := dm.db.Exec("insert into news (text, unixTime) values(?, unix_timestamp(now()))", text)
    
    return err
}

// AddNewsWithTime adds a news displayed on "/" with unixtime
// len(text) <= 256
func (dm *DatabaseManager) AddNewsWithTime(text string, unixTime int) error {
    if len(text) > 256 {
        return errors.New("len(text) > 256")
    }
    
    _, err := dm.db.Exec("insert into news (text, unixTime) values(?, ?)", text, unixTime)
    
    return err
}

// GetNews returns `showedNewCount` recent news of 
func (dm *DatabaseManager) GetNews() ([]News, error) {
    res, err := dm.db.Query("select text, unixTime from news order by unixTime desc limit ?", dm.showedNewCount)
    
    if err != nil {
        return nil, err
    }
    
    var text string
    var unixTime int
    newsSlice := make([]News, 0, dm.showedNewCount)
    
    for res.Next() {
        err := res.Scan(&text, &unixTime)
        
        if err != nil {
            return nil, err
        }
        
        newsSlice = append(newsSlice, News{text, time.Unix(int64(unixTime), 0).String()})
    }
    
    return newsSlice, nil
}