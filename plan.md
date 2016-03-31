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

## Group (groups)
- groupID (int, primary key)
- groupName (varchar(50))

## News (news)
- text (varchar(256))
- unixTime (int, index)

# ToDo
- Implement caches of sessions (session.go)
- Implement /onlinejudge
- Implement /contests