# devops-camp backend

見積もり振り返り型タスク管理ツールのバックエンドAPI。DevOpsキャンプ最終課題 Step3の実装。

## 構成

- 言語/フレームワーク: Go 1.26 / `net/http`(標準ライブラリのみ、フレームワーク不使用)
- データベース: MySQL
- 認証: JWT(`golang-jwt/jwt/v5`)、パスワードは`bcrypt`でハッシュ化
- ディレクトリ構成:
  - `main.go` — エントリポイント、ルーティング定義
  - `internal/config` — 環境変数からの設定読み込み
  - `internal/db` — MySQL接続
  - `internal/models` — `User` / `Task`構造体
  - `internal/auth` — パスワードハッシュ化、JWT発行/検証、認証ミドルウェア
  - `internal/httpx` — JSONレスポンス/エラーレスポンスのヘルパー
  - `internal/handlers` — 各エンドポイントのハンドラ
  - `schema.sql` — テーブル定義(DB仕様書に対応)

## セットアップ

1. ローカルでMySQLを起動し、データベースを作成する

   ```sh
   mysql -u root -e "CREATE DATABASE devops_camp;"
   mysql -u root devops_camp < schema.sql
   ```

2. 依存関係を取得する

   ```sh
   go mod download
   ```

3. 環境変数を設定する(任意、デフォルト値あり)

   | 変数名 | デフォルト値 | 説明 |
   |---|---|---|
   | `PORT` | `8080` | サーバーのリッスンポート |
   | `MYSQL_DSN` | `root@tcp(127.0.0.1:3306)/devops_camp?parseTime=true&charset=utf8mb4` | MySQL接続文字列 |
   | `JWT_SECRET` | `dev-secret-change-me` | JWT署名シークレット(本番相当で使う場合は必ず変更) |

4. サーバーを起動する

   ```sh
   go run .
   ```

## テスト・静的解析

```sh
# 静的解析
go vet ./...
staticcheck ./...   # go install honnef.co/go/tools/cmd/staticcheck@latest で導入

# フォーマットチェック
gofmt -l .

# ユニットテスト(DB不要な部分: JWT / パスワードハッシュ / 認証ミドルウェア)
go test ./...

# 結合テスト(ハンドラ層、実際のMySQLに対して実行。別データベースを推奨)
mysql -u root -e "CREATE DATABASE devops_camp_test;"
TEST_MYSQL_DSN="root@tcp(127.0.0.1:3306)/devops_camp_test?parseTime=true&charset=utf8mb4" go test ./... -v
```

`TEST_MYSQL_DSN`が未設定の場合、DBを要するテストは自動的にスキップされる。

## API動作確認例(curl)

```sh
# ユーザー登録
curl -s -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"taro@example.com","password":"password123","name":"太郎"}'

# ログイン
TOKEN=$(curl -s -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"taro@example.com","password":"password123"}' | jq -r .access_token)

# タスク作成
curl -s -X POST http://localhost:8080/tasks \
  -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d '{"title":"ER図を作成する","estimated_weight":3}'

# タスク一覧
curl -s http://localhost:8080/tasks -H "Authorization: Bearer $TOKEN"

# タスク完了
curl -s -X PATCH http://localhost:8080/tasks/1 \
  -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d '{"status":"done","actual_weight":4}'

# 日別集計(コントリビューションカレンダー用)
curl -s http://localhost:8080/tasks/calendar -H "Authorization: Bearer $TOKEN"
```
