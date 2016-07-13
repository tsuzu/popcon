package main

import "errors"

// Group is a struct to save GroupData
type Group struct {
    Gid int `db:"pk"`
    Name string `default:""`
}

func (dm *DatabaseManager) CreateGroupTable() error {
	err := dm.db.CreateTableIfNotExists(&Group{})

	if err != nil {
		return err
	}

	dm.db.CreateUniqueIndex(&User{}, "nameIdx", "name")

	return nil
}

// GroupAdd adds a new group
// len(groupName) <= 50
func (dm *DatabaseManager) GroupAdd(groupName string) (int64, error) {
    if len(groupName) > 50 {
        return 0, errors.New("len(groupName) > 50")
    }
    
	group := &Group{Name: groupName}

	return dm.db.Insert(&group)
}

// GroupFind finds a group with groupID
func (dm *DatabaseManager) GroupFind(groupID int) (*Group, error) {
    var resulsts []Group

	err := dm.db.Select(&resulsts, dm.db.Where("gid", "=", groupID))
	
	if err != nil {
        return nil, err
    }
    
    if len(resulsts) == 0 {
		return nil, errors.New("Unknown group")
	}

	return &resulsts[0], nil	
}

// GroupRemove removes from groups
func (dm *DatabaseManager) GroupRemove(groupID int) error {
	_, err := dm.db.Delete(&Group{Gid: groupID})

	return err 
}