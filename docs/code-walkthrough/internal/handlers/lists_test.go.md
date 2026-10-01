# internal/handlers/lists_test.go

`lists.go`およびリスト機能に関連する`tasks.go`側の挙動の結合テスト。実際のMySQL(`TEST_MYSQL_DSN`)に対して実行する。

## 共通セットアップ

`newListHandlerWithUser(t)`が`setupTestDB(t)`でDBを初期化し、テスト用ユーザーを1件INSERTしてから`(*handlers.ListHandler, *handlers.TaskHandler, secret []byte, userID int64)`を返す。`TaskHandler`も一緒に返すのは、リスト削除時のタスク側の挙動(`list_id`が`NULL`に戻ること)を検証するため。

## テストケース

| テスト名 | 検証内容 |
|---|---|
| `TestListLifecycle_CreateUpdateDelete` | リスト作成→一覧取得で1件含まれることを確認→リネーム→削除、の一連の流れを検証 |
| `TestListDelete_UnsetsTaskListID` | `list_id`を指定してタスクを作成→そのリストを削除→タスクを再取得して`list_id`が`null`に戻っていることを確認(`ON DELETE SET NULL`の挙動、[ADR-0021](../../../adr/0021-task-list-grouping.md)) |
| `TestTaskCreate_ListIDNotAllowedForChildTask` | 親タスクを作成→その子タスクを`list_id`付きで作成しようとすると`422 LIST_ID_NOT_ALLOWED_FOR_CHILD_TASK`になることを確認 |
| `TestTaskCreate_InvalidListID` | 存在しない`list_id`を指定すると`422`になることを確認 |

## 読むときのポイント

- `TestListDelete_UnsetsTaskListID`は`ListHandler.Delete`と`TaskHandler.Get`をまたいで検証しており、DBの外部キー制約(アプリケーションコードではなくMySQL側)が実際に効いていることを確認する数少ないテストの1つ
- `tasks_test.go`ではなくこのファイルに`list_id`関連のタスク側テストも置いているのは、`lists`テーブルのセットアップ(`newListHandlerWithUser`)に依存するテストをまとめたいため
