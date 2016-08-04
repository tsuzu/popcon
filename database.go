package main

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/naoina/genmai"
	"time"
	//	"os"
)

// Shared in all codes
var mainDB *DatabaseManager

// DatabaseManager is a connector to this database
type DatabaseManager struct {
	db *genmai.DB
	//db *sql.DB
	showedNewCount int
}

// NewDatabaseManager is a function to initialize database connections
// static function
func NewDatabaseManager() (*DatabaseManager, error) {
	dm := &DatabaseManager{}
	var err error

	// pcpjudge Database
	dm.db, err = genmai.New(&genmai.MySQLDialect{}, settings.DB)
	//	dm.db, err = sql.Open("mysql", "popcon:password@/popcon") // Should change password

	if err != nil {
		return nil, err
	}

	dm.db.DB().SetConnMaxLifetime(3 * time.Minute)
	dm.db.DB().SetMaxIdleConns(150)
	dm.db.DB().SetMaxOpenConns(150)
	//dm.db.SetLogOutput(os.Stdout)

	err = dm.db.DB().Ping()

	if err != nil {
		return nil, err
	}

	// user_and_group.go
	// Create Users Table
	err = dm.CreateUserTable()
	//_, err = dm.db.Exec("create table if not exists users (internalID int(11) auto_increment primary key, userID varchar(20) unique, userName varchar(256) unique, passHash varbinary(64), email varchar(50), groupID int(11))")

	if err != nil {
		return nil, err
	}

	// session.go
	// Create Sessions Table
	// TODO: Fix a bug about Year 2038 Bug in unixTimeLimit
	err = dm.CreateSessionTable()
	//_, err = dm.db.Exec("create table if not exists sessions (sessionID varchar(50) primary key, internalID int(11), unixTimeLimit int(11), index iid(internalID), index idx(unixTimeLimit))")

	if err != nil {
		return nil, err
	}

	// user_and_group.go
	err = dm.CreateGroupTable()
	//_, err = dm.db.Exec("create table if not exists groups (groupID int(11) auto_increment primary key, groupName varchar(50))")

	if err != nil {
		return nil, err
	}

	// news.go
	err = dm.CreateNewsTable()
	//_, err = dm.db.Exec("create table if not exists news (text varchar(256), unixTime int, index uti(unixTime))")

	if err != nil {
		return nil, err
	}

	err = dm.CreateContestTable()

	if err != nil {
		return nil, err
	}

	err = dm.CreateContestProblemTable()

	if err != nil {
		return nil, err
	}

	err = dm.CreateSubmissionTable()

	if err != nil {
		return nil, err
	}

	err = dm.CreateContestParticipationTable()

	if err != nil {
		return nil, err
	}

	err = dm.CreateLanguageTable()

	if err != nil {
		return nil, err
	}

	return dm, nil
}
