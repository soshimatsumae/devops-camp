# Step6 課題提出: フロントエンド実装

Step6の課題ページ([final-challenge/app/step6](https://devopscamp.reheartcloud.com/final-challenge/app/step6))に沿って、Step4の画面一覧・Step5の画面設計変更(`docs/assignment/step5.md`)をもとに、実際に動くフロントエンドを実装した。

## 1. 概要

### 技術構成

- フレームワーク: Next.js 16(App Router)+ TypeScript
- スタイリング: Tailwind CSS v4
- アイコン: lucide-react(絵文字は使わず、全てアイコンコンポーネントに統一)
- データ取得: 素の`fetch` + `useState`/`useEffect`(SWR・TanStack Queryのようなライブラリは使わず、学習のしやすさを優先する方針で壁打ちして決定)
- 自動テスト(Jest + React Testing Library): 今回は未導入(課題上は任意。フロントエンドの品質担保は型チェック・ESLintを中心にしている。「3. テスト結果」の内容とは別の話)

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

3コマンドともエラー・警告なしで成功した(`tsc --noEmit`・`eslint`は問題がなければ何も出力しない)。

```
> frontend@0.1.0 lint
> eslint


> frontend@0.1.0 build
> next build

▲ Next.js 16.4.0 (Turbopack)
✓ Running next.config.ts took 1182ms

- Cache Components enabled
- Partial Prefetching enabled

  Creating an optimized production build ...
✓ Compiled successfully in 1411ms
  Running TypeScript ...
  Finished TypeScript in 798ms ...
  Collecting page data using 7 workers ...
✓ Generating static pages using 7 workers (6/6) in 185ms
  Finalizing page optimization ...

Route (app)
┌ ○ /
├ ○ /_not-found
├ ○ /login
└ ○ /register


○  (Static)  prerendered as static content
```

### バックエンド

```sh
go vet ./...
go build ./...
go test ./...

# 結合テスト(別データベースを使用)
TEST_MYSQL_DSN="root:<password>@tcp(127.0.0.1:3306)/devops_camp_test?parseTime=true&charset=utf8mb4" go test -count=1 ./... -v
```

`go vet`・`go build`はエラーなし(問題がなければ何も出力しない)。結合テストは`-count=1`でキャッシュを使わずに実行し、31件すべてPASS(SKIP・FAILなし)。以下は`-v`出力から各テストの結果行を抜き出したもの。

```
--- PASS: TestGenerateAndParseToken (0.00s)
--- PASS: TestParseToken_WrongSecret (0.00s)
--- PASS: TestParseToken_Expired (0.00s)
--- PASS: TestParseToken_Garbage (0.00s)
--- PASS: TestRequireAuth_MissingHeader (0.00s)
--- PASS: TestRequireAuth_ValidToken (0.00s)
--- PASS: TestRequireAuth_InvalidToken (0.00s)
--- PASS: TestHashAndCheckPassword (0.15s)
PASS
ok  	github.com/smatsumae/devops-camp/internal/auth	0.483s
--- PASS: TestParseCalendarRange_Default (0.00s)
--- PASS: TestParseCalendarRange_ExplicitRange (0.00s)
--- PASS: TestParseCalendarRange_InvalidFormat (0.00s)
--- PASS: TestParseCalendarRange_FromAfterTo (0.00s)
--- PASS: TestRegister_Success (0.11s)
--- PASS: TestRegister_DuplicateEmail (0.12s)
--- PASS: TestRegister_ShortPassword (0.03s)
--- PASS: TestLogin_Success (0.11s)
--- PASS: TestLogin_InvalidCredentials (0.13s)
--- PASS: TestListLifecycle_CreateUpdateDelete (0.04s)
--- PASS: TestListDelete_UnsetsTaskListID (0.03s)
--- PASS: TestTaskCreate_ListIDNotAllowedForChildTask (0.02s)
--- PASS: TestTaskCreate_InvalidListID (0.02s)
--- PASS: TestTaskCreate_RequiresEstimatedHours (0.02s)
--- PASS: TestTaskCreate_EstimatedHoursOutOfRange (0.02s)
--- PASS: TestTaskCreate_EstimatedHoursAllowsDecimal (0.03s)
--- PASS: TestTaskCreate_InvalidParentID (0.03s)
--- PASS: TestTaskLifecycle_CreateGetUpdateDelete (0.03s)
--- PASS: TestTaskList_KeywordSearch (0.02s)
--- PASS: TestTaskList_SortByEstimatedHours (0.02s)
--- PASS: TestTaskList_InvalidSort (0.02s)
--- PASS: TestTaskGet_OtherUsersTaskNotFound (0.02s)
--- PASS: TestTaskCalendar_AggregatesByCompletionDate (0.03s)
PASS
ok  	github.com/smatsumae/devops-camp/internal/handlers	1.502s
```

## 4. 振り返り

### 苦労したこと・学んだこと

**ヒートマップに色が反映されないバグ**: カレンダー集計APIが返す日付が、`"2026-10-09"`ではなく`"2026-10-09T00:00:00Z"`というISO日時形式になってしまっていた。MySQLの`DATE`型を、`parseTime=true`の接続設定下でGoの`string`型にスキャンすると、一度`time.Time`を経由してRFC3339形式の文字列に変換されることが原因だった。既存のテストは集計値(`total_hours`)しか検証しておらず、日付の形式までは見ていなかったため見逃されていた。`DATE_FORMAT`でMySQL側から明示的に文字列として返すよう修正し、テストにも日付の形式を検証するアサーションを追加した

### 今後の課題

- 配色・フォント・詳細レイアウトの作り込み(現状はTailwindの標準的な見た目にとどめている)
- Jest + React Testing Libraryによる自動テストの導入(課題上は任意)
- Googleカレンダー連携(Step4から後回しにしたまま未着手)
