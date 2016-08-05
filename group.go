package main

import "errors"

// Group is a struct to save GroupData
type Group struct {
    Gid int64 `db:"pk"`
    Name string `default:""`
}

func (dm *DatabaseManager) CreateGroupTable() error {
	err := dm.db.CreateTableIfNotExists(&Group{})

	if err != nil {
		return err
	}

	dm.db.CreateUniqueIndex(&User{}, "name")

	return nil
}

// GroupAdd adds a new group
// len(groupName) <= 50
func (dm *DatabaseManager) GroupAdd(groupName string) (int64, error) {
    if len(groupName) > 50 {
        return 0, errors.New("len(groupName) > 50")
    }
    
	group := &Group{Name: groupName}

	_, err := dm.db.Insert(&group)

	if err != nil {
		return 0, err
	}

	return group.Gid, nil
}

// GroupFind finds a group with groupID
func (dm *DatabaseManager) GroupFind(gid int) (*Group, error) {
    var resulsts []Group

	err := dm.db.Select(&resulsts, dm.db.Where("gid", "=", gid))
	
	if err != nil {
        return nil, err
    }
    
    if len(resulsts) == 0 {
		return nil, errors.New("Unknown group")
	}

	return &resulsts[0], nil	
}

// GroupRemove removes from groups
func (dm *DatabaseManager) GroupRemove(gid int64) error {
	_, err := dm.db.Delete(&Group{Gid: gid})

	return err 
}