# popcon

## What is popcon?
- popconはオープンソースな競技プログラミング用コンテストサーバです。
- 主に部活内でコンテストを開催するために製作しています。
- メインのコードは全てGo、WebページはBootstrap3を使用。
- **JOI以外のコンテスト形式の実装が終わっていません**

## Dependencies
- [go.uuid](http://github.com/satori/go.uuid)
- [mysql (go-sql-driver)](http://github.com/go-sql-driver/mysql)
- [genmai](https://github.com/naoina/genmai)
- [Blackfriday](https://github.com/russross/blackfriday)
- [Bluemonday](https://github.com/microcosm-cc/bluemonday)
- [Gorilla/Handlers](https://github.com/gorilla/handlers)
- [Gorilla/Websocket](https://github.com/gorilla/websocket)
- MySQL

## How to install
- install MySQL
- Write settings in "popcon.json" (ex. popcon_template.json, autocommit must be off.)
- Run "go get github.com/cs3238-tsuzu/popcon"
- Run "mkdir submissions contest_problems contests"
- Become root
- ./popcon

## License
- Under the MIT License
- Copyright (c) 2016 Tsuzu
