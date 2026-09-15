# devops-camp

DevOpsキャンプ最終課題として開発する、見積もり振り返り型タスク管理ツール。テーマ選定からバックエンド・フロントエンド・インフラ・コンテナ化・CI/CDまで一気通貫で作るリポジトリ。設計判断の背景は [docs/adr/README.md](docs/adr/README.md) を参照。

## 進捗状況
参考：[アプリケーションコースの概要](https://devopscamp.reheartcloud.com/final-challenge/app/overview)

| Step | 内容 | 状態 |
|---|---|---|
| 1-2 | テーマ選定・API/DB仕様書 | 完了 |
| 3 | バックエンド実装 | 完了 |
| 4-6 | フロントエンド(学習・設計・実装) | 未着手 |
| 7-8 | インフラ(学習・設計・クラウド展開) | 未着手 |
| 9-10 | コンテナ化(学習・運用移行) | 未着手 |
| 11 | CI/CDパイプライン | 未着手 |
| 12 | 非機能要件の検討 | 未着手 |

## リポジトリ構成

現時点ではバックエンドのみ。フロントエンド・インフラ関連のコードは対応するStepに進んだ段階で追加していく。

- `main.go` — エントリポイント、ルーティング定義
- `internal/config` — 環境変数からの設定読み込み
- `internal/db` — MySQL接続
- `internal/models` — `User` / `Task`構造体
- `internal/auth` — パスワードハッシュ化、JWT発行/検証、認証ミドルウェア
- `internal/httpx` — JSONレスポンス/エラーレスポンスのヘルパー
- `internal/handlers` — 各エンドポイントのハンドラ
- `schema.sql` — テーブル定義(DB仕様書に対応)

---

## Step 1-2 (テーマ選定・API/DB仕様書)

### テーマ

> **テーマ**
> - タスクごとに大変さの「想定」と「実績」を記録し、見立ての精度を上げられるタスク管理ツール
> - 完了したタスクの重さを日別に集計し、GitHubのコントリビューションカレンダーのようなヒートマップで頑張りを可視化する
>
> **ターゲットユーザー**
> - 自分の作業量やタスクの見積もり精度を把握・改善したい人(僕)
>
> **解決する課題**
> - タスクの見積もりと実際の作業量にギャップがあることが多く、それを振り返る手段がない
> - 日々の頑張りが定量的に見えないため、モチベーションを維持しづらい
>
> **機能ドラフト案**
> - タスクの作成・一覧・更新・削除
> - 子タスク・孫タスクの作成
> - タスク作成時に「想定の大変さ」を入力
> - タスク完了時に「実際の大変さ」を入力
> - 見積もりと実績の差分の記録・振り返り表示
> - 完了タスクの重さを日別に集計したコントリビューションカレンダー表示
> - (できたらいいな)Googleカレンダー連携

### 仕様書

- [API仕様書](./docs/API_SPEC.md)
- [DB仕様書](./docs/DB_SPEC.md)
- [ER図](./docs/ER_DIAGRAM.md)

---

## バックエンド (Step 3)

言語/フレームワーク: Go / `net/http`(標準ライブラリのみ、フレームワーク不使用)、DB: MySQL、認証: JWT(`golang-jwt/jwt/v5`) + `bcrypt`。

### セットアップ

1. ローカルでMySQLを起動し、データベースを作成する(root にパスワードを設定している場合は `-p` を付ける)

   ```sh
   mysql -u root -p -e "CREATE DATABASE devops_camp;"
   mysql -u root -p devops_camp < schema.sql
   ```

2. 依存関係を取得する

   ```sh
   go mod download
   ```

3. 環境変数を設定する(任意、デフォルト値あり)

   | 変数名 | デフォルト値 | 説明 |
   |---|---|---|
   | `PORT` | `8080` | サーバーのリッスンポート |
   | `MYSQL_DSN` | `root@tcp(127.0.0.1:3306)/devops_camp?parseTime=true&charset=utf8mb4` | MySQL接続文字列。root にパスワードを設定している場合は `root:<password>@tcp(...)/...` の形で明示的に指定する(デフォルト値はパスワードなし前提) |
   | `JWT_SECRET` | `dev-secret-change-me` | JWT署名シークレット(本番相当で使う場合は必ず変更) |

4. サーバーを起動する

   ```sh
   go run .
   ```

### テスト・静的解析

```sh
# 静的解析
go vet ./...
staticcheck ./...   # go install honnef.co/go/tools/cmd/staticcheck@latest で導入

# フォーマットチェック
gofmt -l .

# ユニットテスト(DB不要な部分: JWT / パスワードハッシュ / 認証ミドルウェア)
go test ./...

# 結合テスト(ハンドラ層、実際のMySQLに対して実行。別データベースを推奨)
# root にパスワードを設定している場合は DSN に root:<password>@... の形で含める
mysql -u root -p -e "CREATE DATABASE devops_camp_test;"
TEST_MYSQL_DSN="root:<password>@tcp(127.0.0.1:3306)/devops_camp_test?parseTime=true&charset=utf8mb4" go test ./... -v
```

`TEST_MYSQL_DSN`が未設定の場合、DBを要するテストは自動的にスキップされる。

### API動作確認例(curl)

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

## フロントエンド (Step 4-6)

未着手。着手時にこのセクションへセットアップ手順を追記する。

## インフラ・コンテナ・CI/CD (Step 7-11)

未着手。着手時にこのセクションへデプロイ手順・コンテナ起動手順・CI/CDパイプラインの説明を追記する。
