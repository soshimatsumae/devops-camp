# internal/handlers/testhelpers_test.go

`internal/handlers`パッケージの結合テストが共通で使うヘルパー集。

## やっていること

- `setupTestDB(t) *sql.DB`
  - `TEST_MYSQL_DSN`環境変数が未設定なら`t.Skip`でテスト自体をスキップ(ローカルMySQLが無い環境やこのサンドボックスから実行してもテストが失敗しない)
  - 接続後、`schema.sql`を`;`区切りで読み込んで実行し、スキーマを再適用
  - `TRUNCATE TABLE tasks` / `TRUNCATE TABLE users`でテーブルを空にする(`FOREIGN_KEY_CHECKS`を一時的に無効化してから実行し、外部キー制約でTRUNCATEが失敗しないようにしている)
  - `t.Cleanup`で接続クローズを予約
- `newAuthedRequest(t, secret, userID, method, target, body) *http.Request`
  - `httptest.NewRequest`でリクエストを組み立て、`auth.GenerateToken`で発行したJWTを`Authorization`ヘッダーにセット
  - さらに`auth.ContextWithUserID`で**コンテキストにも直接**ユーザーIDを注入する
- `decodeJSON(t, rec, v)`: `httptest.ResponseRecorder`のボディをJSONデコードする共通処理

## 読むときのポイント

- `newAuthedRequest`がヘッダーとコンテキストの**両方**にユーザー情報をセットしているのが重要なポイント。テストはハンドラを`auth.RequireAuth`ミドルウェア経由ではなく直接呼び出すため、ヘッダーを設定するだけでは`internal/auth.UserIDFromContext`が何も返さず、ハンドラは`401 UNAUTHORIZED`を返してしまう。実際にこの配線が漏れていたことでタスク関連の結合テストが全滅した経緯があり、`auth.ContextWithUserID`(`internal/auth/middleware.go`)はこの問題を修正するために追加されたexported関数
- ヘッダーへのトークン設定自体は(コンテキスト注入がある以上)テストの通過には必須ではないが、実際のクライアントに近いリクエストの形を保つため、また将来ミドルウェア経由のテストに切り替えても動くようにするために残している
- `setupTestDB`が毎テストでスキーマ再適用+TRUNCATEを行うため、テスト間の状態リークがなく、実行順序に依存しないテストになっている
