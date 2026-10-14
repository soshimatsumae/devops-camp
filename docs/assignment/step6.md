# Step6 課題提出: フロントエンド実装

Step6の課題ページ([final-challenge/app/step6](https://devopscamp.reheartcloud.com/final-challenge/app/step6))に沿って、Step4の画面一覧・Step5の画面設計変更(`docs/assignment/step5.md`)をもとに、実際に動くフロントエンドを実装した。

## 1. 概要

### 技術構成

- フレームワーク: Next.js 16(App Router)+ TypeScript
- スタイリング: Tailwind CSS v4
- アイコン: lucide-react(絵文字は使わず、全てアイコンコンポーネントに統一)
- データ取得: 素の`fetch` + `useState`/`useEffect`(SWR・TanStack Queryのようなライブラリは使わず、学習のしやすさを優先する方針で壁打ちして決定)
- テスト: 今回は未導入(課題上は任意。型チェック・ESLintを品質担保の中心にしている)

コードは`frontend/`配下に追加し、バックエンド(Go)とは別のディレクトリ・別の`package.json`で管理している。`frontend/README.md`にセットアップ手順を記載。

### 実装のポイント

- **認証**: `POST /auth/login`で受け取った`access_token`を`localStorage`に保存(`lib/auth.ts`)。`lib/api.ts`の`fetch`ラッパーが全リクエストに`Authorization: Bearer`ヘッダーを自動付与する。認証必須のAPIが`401`を返したら、トークンを破棄してS-01へ強制リダイレクトする(ログイン自体の`401 INVALID_CREDENTIALS`は対象外にして、通常のエラーとして画面内に表示する)
- **画面構成**: S-01(ログイン)・S-02(新規登録)・S-03(ダッシュボード)はNext.jsのルーティングで切り替え、M-01〜M-04(タスク作成・編集・詳細・リスト管理)はページ遷移を伴わないモーダルとして、React stateで開閉を管理する
- **タスク一覧のグループ表示**: `docs/assignment/step5.md`「7.」の設計通り、「リスト: すべて」のときはクライアント側で`list_id`ごとにグルーピングして表示(ページネーションなし)、特定のリストを選んだときはサーバー側のページネーション付きフラット表示に切り替える
- **子・孫タスクの展開**: ネイティブの`<details>`/`<summary>`要素でアコーディオンを実装。クリックのたびに`GET /tasks?parent_id=`で取得し、一度取得した子タスクはキャッシュして再展開時に取得し直さない。キーボード操作(Tab/Enter/Space)はブラウザ標準の挙動をそのまま使っている
- **バックエンド側の変更**: 別オリジン(Next.jsの開発サーバー)からのアクセスを許可するCORSミドルウェア(`internal/httpx/cors.go`)を追加した

## 2. ソースコード

GitHubリポジトリ: https://github.com/soshimatsumae/devops-camp

## 3. テスト結果

### フロントエンド

```sh
cd frontend
npx tsc --noEmit   # 型チェック
npm run lint       # ESLint
npm run build      # 本番ビルド
```

<!-- TODO: 上記3コマンドの実行結果(エラー・警告が出ていないこと)を貼る -->

### バックエンド

```sh
go vet ./...
go build ./...
go test ./...

# 結合テスト(別データベースを使用)
TEST_MYSQL_DSN="root:<password>@tcp(127.0.0.1:3306)/devops_camp_test?parseTime=true&charset=utf8mb4" go test ./... -v
```

<!-- TODO: 上記コマンドの実行結果を貼る -->

## 4. 振り返り

### 苦労したこと・学んだこと

**ヒートマップに色が反映されないバグ**: カレンダー集計APIが返す日付が、`"2026-10-09"`ではなく`"2026-10-09T00:00:00Z"`というISO日時形式になってしまっていた。MySQLの`DATE`型を、`parseTime=true`の接続設定下でGoの`string`型にスキャンすると、一度`time.Time`を経由してRFC3339形式の文字列に変換されることが原因だった。既存のテストは集計値(`total_hours`)しか検証しておらず、日付の形式までは見ていなかったため見逃されていた。`DATE_FORMAT`でMySQL側から明示的に文字列として返すよう修正し、テストにも日付の形式を検証するアサーションを追加した

### 今後の課題

- 配色・フォント・詳細レイアウトの作り込み(現状はTailwindの標準的な見た目にとどめている)
- Jest + React Testing Libraryによる自動テストの導入(課題上は任意)
- Googleカレンダー連携(Step4から後回しにしたまま未着手)
