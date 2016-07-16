# Table(genmaiに変更したからいらなくなった)

## User (users)
- Iid (int, auto increment, primary key)
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
- if not existsがないからcreate indexがコケる可能性がある(無視で大丈夫そう?)
- submissionのDB操作を実装

# contestのフォルダ構成
- ./contest_problems
    - XXX(pid)/
        - prob ... 問題文(サンプルを含む)のhtml or Markdown
        - .cases_lock ... casesのロック
        - cases/
            XXX_in ... 入力
            XXX_out ...出力（複数
            data ... テストケースの名前及びテストケースとスコアの紐付け

# submissionsのフォルダ構成
- ./submissions/
    - XXX(sid)
        - code ... ソースコード
        - msg ... コンパイルエラーとかInternalErrorのメッセージ
        - case ... 各ケースの結果のjson情報