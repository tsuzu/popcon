package main

// type ContestRanking struct {
//     Cid int64 `default:""`
//     Iid int64 `default:""`
//     Score int64 `default:""`
//     Time int64 `default:""`
//     Details string `default:"" size:"5000"`
// }

// func (dm *DatabaseManager) CreateContestRankingTable() error {
//     err := dm.db.CreateTableIfNotExists(&ContestRanking{})

//     if err != nil {
//         return err
//     }

//     dm.db.CreateIndex(&ContestRanking{}, "cid")
//     dm.db.CreateIndex(&ContestRanking{}, "iid")
    
// }