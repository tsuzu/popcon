# Table

## User (users)
- internalID (int, auto increment, primary key)
- userID (varchar(20), unique index)
- userName (varchar(256), unique index)
- passHash (varbinary(64), \[64\]uint8)
- email (varchar(50))

## Session (sessions)
- sessionID (varchar(50), primary key)
- internalID (int, index)
- unixTimeLimit(int, index)

# ToDo
- Divide codes in multiple files
- Implement caches of sessions (session.go)