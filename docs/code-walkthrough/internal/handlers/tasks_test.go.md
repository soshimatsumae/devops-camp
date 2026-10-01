# internal/handlers/tasks_test.go

`tasks.go`の結合テスト。実際のMySQL(`TEST_MYSQL_DSN`)に対して実行する。

## 共通セットアップ

`newTaskHandlerWithUser(t)`が`setupTestDB(t)`でDBを初期化し、テスト用ユーザーを1件INSERTしてから`(*handlers.TaskHandler, secret []byte, userID int64)`を返す。各テストはこれを起点にする。

## テストケース

| テスト名 | 検証内容 |
|---|---|
| `TestTaskCreate_RequiresEstimatedHours` | `estimated_hours`なしの作成リクエストは`422 ESTIMATED_HOURS_REQUIRED` |
| `TestTaskCreate_EstimatedHoursOutOfRange` | `estimated_hours`が最小値(0.5)未満だと`422` |
| `TestTaskCreate_EstimatedHoursAllowsDecimal` | `estimated_hours`に`1.5`のような小数を指定できることを確認 |
| `TestTaskCreate_InvalidParentID` | 存在しない`parent_id`を指定すると`422`(`INVALID_PARENT_ID`) |
| `TestTaskLifecycle_CreateGetUpdateDelete` | 親タスク作成→子タスク作成→`GET`で子タスクが`children`に含まれることを確認→`actual_hours`なしで`done`にしようとして`422`→`actual_hours`ありで`done`に更新して`completed_at`が設定されることを確認→再取得して永続化を確認→削除→削除後の`GET`が`404`になることを一通り検証する、最も長いテスト |
| `TestTaskList_KeywordSearch` | `q`パラメータでtitle/descriptionへの部分一致検索ができることを確認 |
| `TestTaskList_SortByEstimatedHours` | `sort=estimated_hours_asc`で見積もり工数の昇順に並び替えられることを確認 |
| `TestTaskList_InvalidSort` | ホワイトリストにない`sort`値は`400`になることを確認 |
| `TestTaskGet_OtherUsersTaskNotFound` | 別ユーザー(`jiro`)を追加で作成し、taroのタスクをjiroとして取得しようとすると`404`になることを確認(所有者チェック) |
| `TestTaskCalendar_AggregatesByCompletionDate` | 工数3時間と5時間のタスクを作成して両方完了させ、`GET /tasks/calendar`で同日の合計(`8`)に集計されることを確認 |

## 読むときのポイント

- `newAuthedRequest(t, secret, userID, ...)`(`testhelpers_test.go`)でリクエストを組み立てているが、ハンドラを`RequireAuth`ミドルウェア経由ではなく直接呼んでいるため、Authorizationヘッダーだけでなく`auth.ContextWithUserID`でコンテキストにも直接ユーザーIDを注入している。ここを見落とすと全テストが`401 UNAUTHORIZED`で落ちる(実際に一度このバグで全滅した経緯がある)
- `TestTaskLifecycle_CreateGetUpdateDelete`は1つのテスト内で作成・参照・更新・削除の一連の流れを追っており、単体テストというよりシナリオテストに近い。個別の関数の正しさよりも「実際にAPIとして使ったときに壊れていないか」を見るために意図的にまとめている
- `TestTaskGet_OtherUsersTaskNotFound`は`403 Forbidden`ではなく`404`を期待している。これは`tasks.go`の`loadTask`が所有者チェックを兼ねる設計([tasks.go](./tasks.go.md)参照)に対応するテスト
