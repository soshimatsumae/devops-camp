# main.go

エントリポイント。アプリの起動処理とルーティング定義のみを担う。

## やっていること

1. `config.Load()`で環境変数から設定を読み込む
2. `db.Open(cfg.MySQLDSN)`でMySQLに接続(接続確認まで含む。失敗したら`log.Fatalf`で即終了)
3. `AuthHandler`・`TaskHandler`をDB接続とJWTシークレット・トークンTTLを渡して生成
4. `auth.RequireAuth(cfg.JWTSecret)`で認証デコレータを作り、`/tasks`系ハンドラをラップ
5. `http.NewServeMux()`(Go 1.22+の拡張ServeMux)にメソッド+パスパターンでルートを登録
6. `http.ListenAndServe`でサーバー起動

## ルーティング一覧

| メソッド | パス | 認証 | ハンドラ |
|---|---|---|---|
| POST | `/auth/register` | 不要 | `authHandler.Register` |
| POST | `/auth/login` | 不要 | `authHandler.Login` |
| GET | `/tasks` | 必要 | `taskHandler.List` |
| POST | `/tasks` | 必要 | `taskHandler.Create` |
| GET | `/tasks/calendar` | 必要 | `taskHandler.Calendar` |
| GET | `/tasks/{id}` | 必要 | `taskHandler.Get` |
| PATCH | `/tasks/{id}` | 必要 | `taskHandler.Update` |
| DELETE | `/tasks/{id}` | 必要 | `taskHandler.Delete` |

## 読むときのポイント

- `GET /tasks/calendar`は`GET /tasks/{id}`より前に登録されているが、これは`http.ServeMux`が「より具体的なパターンを優先する」ため実際は順序に依存しない(`calendar`はリテラル一致、`{id}`はワイルドカード)。それでも紛らわしさを避けるため具体的なパスを先に書いている
- ミドルウェア(`requireAuth`)は関数を受け取って関数を返すデコレータパターンで、フレームワークを使わずに横断的関心事(認証)を挟んでいる([ADR-0002](../adr/0002-net-http-only.md))
- `main`関数がDI(依存性注入)コンテナも兼ねている。テストではこの配線を経由せず、各ハンドラを直接生成している([testhelpers_test.go](./internal/handlers/testhelpers_test.go.md)参照)
