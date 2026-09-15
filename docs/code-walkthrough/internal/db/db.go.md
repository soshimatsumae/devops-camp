# internal/db/db.go

MySQL接続の生成のみを担当する、1関数だけの小さなパッケージ。

## やっていること

- `Open(dsn string) (*sql.DB, error)`
  1. `sql.Open("mysql", dsn)`でコネクションプールのハンドルを作成(この時点では実際には接続しない、`database/sql`の仕様)
  2. `conn.Ping()`で実際に接続できるか確認
  3. `Ping`が失敗したら`conn.Close()`してからエラーを返す(コネクションリークを防ぐ)
- `_ "github.com/go-sql-driver/mysql"`のブランクインポートでMySQLドライバを`database/sql`に登録している

## 読むときのポイント

- `sql.Open`はエラーを返さないことが多い(DSNの構文エラー程度しか検知しない)。そのため`Ping()`まで含めて初めて「本当に繋がるか」が分かる。`main.go`が起動時に`db.Open`のエラーを`log.Fatalf`で扱っているのは、この`Ping`失敗を早期に検知するため
- ドライバ名`"mysql"`は`go-sql-driver/mysql`パッケージの`init()`が登録する文字列で、`database/sql`のドライバレジストリを介した疎結合な設計になっている
