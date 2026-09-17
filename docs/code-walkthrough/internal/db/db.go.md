# internal/db/db.go

MySQL接続の生成のみを担当する、1関数だけの小さなパッケージ。

## やっていること

- `Open(dsn string) (*sql.DB, error)`
  1. `sql.Open("mysql", dsn)`でコネクションプールのハンドルを作成(この時点では実際には接続しない、`database/sql`の仕様)
  2. `SetMaxOpenConns(25)`・`SetMaxIdleConns(25)`・`SetConnMaxLifetime(5*time.Minute)`でプールの挙動を明示的に設定
  3. `conn.Ping()`で実際に接続できるか確認
  4. `Ping`が失敗したら`conn.Close()`してからエラーを返す(コネクションリークを防ぐ)
- `_ "github.com/go-sql-driver/mysql"`のブランクインポートでMySQLドライバを`database/sql`に登録している

## 読むときのポイント

- `sql.Open`はエラーを返さないことが多い(DSNの構文エラー程度しか検知しない)。そのため`Ping()`まで含めて初めて「本当に繋がるか」が分かる。`main.go`が起動時に`db.Open`のエラーを`log.Fatalf`で扱っているのは、この`Ping`失敗を早期に検知するため
- ドライバ名`"mysql"`は`go-sql-driver/mysql`パッケージの`init()`が登録する文字列で、`database/sql`のドライバレジストリを介した疎結合な設計になっている
- 接続プールの設定(`MaxOpenConns`等)を明示していないと、`database/sql`のデフォルト挙動に任せることになる。具体的には、MySQL側が長時間アイドルな接続を勝手に切断した後、Goのプールがそれに気づかず古い接続を掴んでエラーになる、といった問題が起きうる。`ConnMaxLifetime`で一定時間ごとに接続を作り直すことでこれを予防している。数値(25/25/5分)は固定値としてハードコードしており、環境ごとに変えたくなったら環境変数化を検討する
