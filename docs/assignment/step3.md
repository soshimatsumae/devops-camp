# Step3 課題提出: バックエンド実装

## 概要

タスク管理ツールのバックエンドAPI。タスクごとに「想定の大変さ」と「実際の大変さ」を記録し、見立ての精度を振り返れること、完了タスクの重さを日別に集計してコントリビューションカレンダー風に可視化できることを目的としている。

- 言語: Go
- フレームワーク: 標準ライブラリ`net/http`のみ(Step1・2で選択した通り、Ginなどのサードパーティフレームワークは不使用。Go 1.22以降の拡張`http.ServeMux`でメソッド+パスパターンによるルーティングを実装)
- データベース: MySQL(`users`/`tasks`の2テーブル。`tasks.parent_id`の自己参照で子・孫タスクの階層を表現)
- 認証: JWT(`golang-jwt/jwt/v5`)。`Authorization: Bearer <token>`ヘッダーを自前実装のミドルウェア(`internal/auth.RequireAuth`)で検証し、リクエストコンテキストへユーザーIDを注入する方式
- パスワード: `bcrypt`でハッシュ化

工夫した点:

- エラーレスポンスを`{"error":{"code":"...","message":"..."}}`の形式に統一し、フロントエンド側の分岐実装をしやすくした
- 他ユーザーのタスクへのアクセスは`403`ではなく`404`を返す設計にし、リソースの存在自体を他ユーザーに知らせないようにした
- `GET /tasks/calendar`は当初パラメータなしで全期間を返す設計だったが、レビューを経て`from`/`to`のクエリパラメータを追加し、未指定時は直近365日をデフォルト範囲とすることでレスポンスサイズの肥大化を防いだ
- 設計上の意思決定はADR(Architecture Decision Record)として`docs/adr/`に記録し、後から「なぜこの設計にしたか」を追えるようにした

## ソースコード

<https://github.com/soshimatsumae/devops-camp>

## API の動作確認

サーバーは`go run .`で起動(`http://localhost:8080`)。以下、API仕様書([docs/API_SPEC.md](../API_SPEC.md))のエンドポイントごとに実行結果を示す。

### AUTH-01: ユーザー登録 (`POST /auth/register`)

```sh
curl -s -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"jiro@example.com","password":"password123","name":"次郎"}'
```

```json
{"id":3,"email":"jiro@example.com","name":"次郎","created_at":"2026-09-15T15:24:54Z"}
```

### AUTH-02: ログイン (`POST /auth/login`)

```sh
curl -s -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"taro@example.com","password":"password123"}'
```

```json
{"access_token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxIiwiZXhwIjoxNzg5NDU3MTIzLCJpYXQiOjE3ODk0NTM1MjN9.o5tZrrvqcNEQ5LoCa06fBmBIGPVqR0al7Bji26oAumQ","token_type":"Bearer","expires_in":3600}
```

以降のエンドポイントはこのレスポンスの`access_token`を`$TOKEN`として使用する。

### TASK-02: タスク作成 (`POST /tasks`)

```sh
curl -s -X POST http://localhost:8080/tasks \
  -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d '{"title":"ER図を作成する","estimated_weight":3}'
```

```json
{"id":3,"parent_id":null,"title":"ER図を作成する","description":null,"status":"todo","estimated_weight":3,"actual_weight":null,"due_date":null,"completed_at":null,"created_at":"2026-09-15T12:43:19Z","updated_at":"2026-09-15T12:43:19Z"}
```

### TASK-01: タスク一覧取得 (`GET /tasks`)

作成が永続化されていることを一覧取得で確認(このユーザーが作成した3件が表示されている)。

```sh
curl -s http://localhost:8080/tasks -H "Authorization: Bearer $TOKEN"
```

```json
{"data":[{"id":3,"parent_id":null,"title":"ER図を作成する","description":null,"status":"todo","estimated_weight":3,"actual_weight":null,"due_date":null,"completed_at":null,"created_at":"2026-09-15T12:43:19Z","updated_at":"2026-09-15T12:43:19Z"},{"id":2,"parent_id":null,"title":"ER図を作成する","description":null,"status":"todo","estimated_weight":3,"actual_weight":null,"due_date":null,"completed_at":null,"created_at":"2026-09-15T11:07:54Z","updated_at":"2026-09-15T11:07:54Z"},{"id":1,"parent_id":null,"title":"ER図を作成する","description":null,"status":"done","estimated_weight":3,"actual_weight":4,"due_date":null,"completed_at":"2026-09-15T02:08:06Z","created_at":"2026-09-15T11:07:49Z","updated_at":"2026-09-15T11:08:05Z"}],"meta":{"page":1,"per_page":20,"total_count":3,"total_pages":1}}
```

`per_page`を指定して1件ずつページ送りできることも確認(この時点でタスクは3件)。

```sh
curl -s "http://localhost:8080/tasks?per_page=1&page=1" -H "Authorization: Bearer $TOKEN"
curl -s "http://localhost:8080/tasks?per_page=1&page=2" -H "Authorization: Bearer $TOKEN"
```

```json
{"data":[{"id":4,"parent_id":null,"title":"異常系テスト用タスク","description":null,"status":"todo","estimated_weight":2,"actual_weight":null,"due_date":null,"completed_at":null,"created_at":"2026-09-17T11:29:29Z","updated_at":"2026-09-17T11:29:29Z"}],"meta":{"page":1,"per_page":1,"total_count":3,"total_pages":3}}
{"data":[{"id":2,"parent_id":null,"title":"ER図を作成する","description":null,"status":"todo","estimated_weight":3,"actual_weight":null,"due_date":null,"completed_at":null,"created_at":"2026-09-15T11:07:54Z","updated_at":"2026-09-15T11:07:54Z"}],"meta":{"page":2,"per_page":1,"total_count":3,"total_pages":3}}
```

`page=1`と`page=2`で異なる1件ずつが返り、`meta.total_pages`が3(タスク3件 ÷ per_page 1)になっていることから、ページネーションが正しく機能していることを確認できる。

### TASK-04: タスク更新 (`PATCH /tasks/{id}`)

id=3を完了状態に更新する。

```sh
curl -s -X PATCH http://localhost:8080/tasks/3 \
  -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d '{"status":"done","actual_weight":4}'
```

```json
{"id":3,"parent_id":null,"title":"ER図を作成する","description":null,"status":"done","estimated_weight":3,"actual_weight":4,"due_date":null,"completed_at":"2026-09-15T06:23:53Z","created_at":"2026-09-15T12:43:19Z","updated_at":"2026-09-15T15:23:52Z"}
```

### TASK-03: タスク詳細取得 (`GET /tasks/{id}`)

更新後に再取得し、`status`・`actual_weight`・`completed_at`が永続化されていることを確認。

```sh
curl -s http://localhost:8080/tasks/3 -H "Authorization: Bearer $TOKEN"
```

```json
{"actual_weight":4,"children":[],"completed_at":"2026-09-15T06:23:53Z","created_at":"2026-09-15T12:43:19Z","description":null,"due_date":null,"estimated_weight":3,"id":3,"parent_id":null,"status":"done","title":"ER図を作成する","updated_at":"2026-09-15T15:23:52Z"}
```

### TASK-06: 日毎に完了したタスク合計量取得 (`GET /tasks/calendar`)

id=3の完了により、同日に完了した他タスク(重さ4)と合算されて`total_weight: 8`になっていることを確認。

```sh
curl -s http://localhost:8080/tasks/calendar -H "Authorization: Bearer $TOKEN"
```

```json
{"data":[{"date":"2026-09-15T00:00:00Z","total_weight":8}]}
```

Step2レビューで追加した`from`/`to`を指定した場合の動作も確認([ADR-0010](../adr/0010-calendar-date-range-params.md))。

```sh
curl -s "http://localhost:8080/tasks/calendar?from=2026-09-01&to=2026-09-30" \
  -H "Authorization: Bearer $TOKEN"
```

```json
{"data":[{"date":"2026-09-15T00:00:00Z","total_weight":4}]}
```

指定した期間(9月1日〜9月30日)に絞り込まれ、この時点で完了していたタスク1件分(重さ4)の集計が返っている。

### TASK-05: タスク削除 (`DELETE /tasks/{id}`)

```sh
curl -s -X DELETE http://localhost:8080/tasks/3 -H "Authorization: Bearer $TOKEN" -w "\nHTTP status: %{http_code}\n"
```

```
HTTP status: 204
```

削除後に同じidを取得すると`404`になり、永続化(削除)されていることを確認。

```sh
curl -s http://localhost:8080/tasks/3 -H "Authorization: Bearer $TOKEN" -w "\nHTTP status: %{http_code}\n"
```

```json
{"error":{"code":"TASK_NOT_FOUND","message":"task not found"}}
```

```
HTTP status: 404
```

## 異常系の動作確認

正常系だけでなく、想定するエラーレスポンスが実際に返ることも確認した。対応するユニット/結合テスト(`TestRegister_DuplicateEmail`, `TestTaskGet_OtherUsersTaskNotFound`, `TestParseCalendarRange_FromAfterTo`)でも同じシナリオを担保している。

### AUTH-01異常系: 既に登録済みのメールアドレスで登録

```sh
curl -s -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"taro@example.com","password":"password123","name":"太郎"}' \
  -w "\nHTTP status: %{http_code}\n"
```

```json
{"error":{"code":"EMAIL_ALREADY_REGISTERED","message":"email already registered"}}

HTTP status: 422
```

### TASK-03異常系: 他ユーザーのタスクを取得しようとする

taro(userIDが異なる別ユーザー)が作成したid=4のタスクを、jiroとしてログインした状態で取得しようとする。

```sh
curl -s http://localhost:8080/tasks/4 -H "Authorization: Bearer $JIRO_TOKEN" -w "\nHTTP status: %{http_code}\n"
```

```json
{"error":{"code":"TASK_NOT_FOUND","message":"task not found"}}

HTTP status: 404
```

他ユーザーのタスクであっても`403 Forbidden`ではなく`404`を返す設計(タスクの存在自体を教えない)のため、このレスポンスになる。

### TASK-06異常系: fromがtoより後の日付になっている

```sh
curl -s "http://localhost:8080/tasks/calendar?from=2026-02-01&to=2026-01-01" \
  -H "Authorization: Bearer $TOKEN" -w "\nHTTP status: %{http_code}\n"
```

```json
{"error":{"code":"INVALID_DATE_RANGE","message":"from must not be after to"}}

HTTP status: 422
```

## 静的解析・自動テスト

### 静的解析

```sh
$ go vet ./...
$ staticcheck ./...
```

いずれも出力なし(問題なし)。

### フォーマットチェック

```sh
$ gofmt -l .
```

出力なし(フォーマット崩れなし)。

### ユニットテスト(DB不要)

```sh
$ go test ./...
?       github.com/smatsumae/devops-camp        [no test files]
ok      github.com/smatsumae/devops-camp/internal/auth  0.575s
?       github.com/smatsumae/devops-camp/internal/config        [no test files]
?       github.com/smatsumae/devops-camp/internal/db    [no test files]
ok      github.com/smatsumae/devops-camp/internal/handlers      0.707s
?       github.com/smatsumae/devops-camp/internal/httpx [no test files]
?       github.com/smatsumae/devops-camp/internal/models        [no test files]
```

### 結合テスト(実際のMySQLに対して実行)

```sh
$ mysql -u root -p -e "CREATE DATABASE IF NOT EXISTS devops_camp_test;"
$ export TEST_MYSQL_DSN="root:<password>@tcp(127.0.0.1:3306)/devops_camp_test?parseTime=true&charset=utf8mb4"
$ go test ./... -v
```

```
?       github.com/smatsumae/devops-camp        [no test files]
=== RUN   TestGenerateAndParseToken
--- PASS: TestGenerateAndParseToken (0.00s)
=== RUN   TestParseToken_WrongSecret
--- PASS: TestParseToken_WrongSecret (0.00s)
=== RUN   TestParseToken_Expired
--- PASS: TestParseToken_Expired (0.00s)
=== RUN   TestParseToken_Garbage
--- PASS: TestParseToken_Garbage (0.00s)
=== RUN   TestRequireAuth_MissingHeader
--- PASS: TestRequireAuth_MissingHeader (0.00s)
=== RUN   TestRequireAuth_ValidToken
--- PASS: TestRequireAuth_ValidToken (0.00s)
=== RUN   TestRequireAuth_InvalidToken
--- PASS: TestRequireAuth_InvalidToken (0.00s)
=== RUN   TestHashAndCheckPassword
--- PASS: TestHashAndCheckPassword (0.16s)
PASS
ok      github.com/smatsumae/devops-camp/internal/auth  (cached)
?       github.com/smatsumae/devops-camp/internal/config        [no test files]
?       github.com/smatsumae/devops-camp/internal/db    [no test files]
=== RUN   TestParseCalendarRange_Default
--- PASS: TestParseCalendarRange_Default (0.00s)
=== RUN   TestParseCalendarRange_ExplicitRange
--- PASS: TestParseCalendarRange_ExplicitRange (0.00s)
=== RUN   TestParseCalendarRange_InvalidFormat
--- PASS: TestParseCalendarRange_InvalidFormat (0.00s)
=== RUN   TestParseCalendarRange_FromAfterTo
--- PASS: TestParseCalendarRange_FromAfterTo (0.00s)
=== RUN   TestRegister_Success
--- PASS: TestRegister_Success (0.11s)
=== RUN   TestRegister_DuplicateEmail
--- PASS: TestRegister_DuplicateEmail (0.13s)
=== RUN   TestRegister_ShortPassword
--- PASS: TestRegister_ShortPassword (0.03s)
=== RUN   TestLogin_Success
--- PASS: TestLogin_Success (0.12s)
=== RUN   TestLogin_InvalidCredentials
--- PASS: TestLogin_InvalidCredentials (0.12s)
=== RUN   TestTaskCreate_RequiresEstimatedWeight
--- PASS: TestTaskCreate_RequiresEstimatedWeight (0.02s)
=== RUN   TestTaskCreate_InvalidParentID
--- PASS: TestTaskCreate_InvalidParentID (0.02s)
=== RUN   TestTaskLifecycle_CreateGetUpdateDelete
--- PASS: TestTaskLifecycle_CreateGetUpdateDelete (0.02s)
=== RUN   TestTaskGet_OtherUsersTaskNotFound
--- PASS: TestTaskGet_OtherUsersTaskNotFound (0.01s)
=== RUN   TestTaskCalendar_AggregatesByCompletionDate
--- PASS: TestTaskCalendar_AggregatesByCompletionDate (0.02s)
PASS
ok      github.com/smatsumae/devops-camp/internal/handlers      (cached)
?       github.com/smatsumae/devops-camp/internal/httpx [no test files]
?       github.com/smatsumae/devops-camp/internal/models        [no test files]
```
