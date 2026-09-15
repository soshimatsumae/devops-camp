# internal/handlers/auth_test.go

`auth.go`の結合テスト。実際のMySQL(`TEST_MYSQL_DSN`)に対して実行する([testhelpers_test.go](./testhelpers_test.go.md)参照)。

## テストケース

| テスト名 | 検証内容 |
|---|---|
| `TestRegister_Success` | 正常な登録リクエストが`201`を返し、レスポンスに`email`と`id`が含まれる |
| `TestRegister_DuplicateEmail` | 同じメールアドレスで2回登録すると2回目は`422 EMAIL_ALREADY_REGISTERED` |
| `TestRegister_ShortPassword` | 8文字未満のパスワードは`422`(バリデーションエラー) |
| `TestLogin_Success` | 登録直後のユーザーでログインすると`200`、`access_token`と`token_type: "Bearer"`が返る |
| `TestLogin_InvalidCredentials` | 登録済みユーザーに誤ったパスワードでログインすると`401` |

## 読むときのポイント

- `newAuthHandler(t)`が`setupTestDB(t)`(`testhelpers_test.go`)を使ってテストごとにスキーマ再適用・テーブルTRUNCATEを行うため、各テストは独立したまっさらな状態から始まる
- ハンドラを`RequireAuth`ミドルウェア経由ではなく直接呼び出している。`/auth/register`・`/auth/login`はそもそも認証不要のエンドポイントなので、`tasks_test.go`のようにコンテキストへユーザーIDを注入する必要がない
- `TestLogin_Success`/`TestLogin_InvalidCredentials`はテスト内で`Register`を呼んでユーザーを用意してから`Login`を検証しており、DBに直接INSERTするのではなく実際のAPIフローに沿ってセットアップしている
