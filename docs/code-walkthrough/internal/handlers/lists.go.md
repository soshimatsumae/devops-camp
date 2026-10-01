# internal/handlers/lists.go

`ListHandler`が`/lists`系4エンドポイント(CRUD)を実装する。タスクを横断的にグルーピングする「リスト」機能([ADR-0021](../../../adr/0021-task-list-grouping.md))のバックエンド側。

## 共通部品

- `scanList(row)`・`listColumns`: `tasks.go`の`scanTask`/`taskColumns`と同じパターン。`List`構造体にNULL許容カラムがないため、`sql.NullXxx`を介さずそのままスキャンできる
- `parseListID(r)`: `tasks.go`の`parseTaskID`と同じ、パスパラメータ`{id}`のパース

## エンドポイントごとの実装

### `List` (`GET /lists`)

- `WHERE user_id = ?`のみ。ページネーションはせず、作成日時の昇順で全件返す(個人利用でリスト件数が大きくならない想定のため、[ADR-0021](../../../adr/0021-task-list-grouping.md))

### `Create` (`POST /lists`)

- `name`必須(トリム後空文字はNG)
- 作成後は直接`SELECT`し直してから返す(`tasks.go`の`Create`が`loadTask`で読み直すのと同じ考え方)

### `Update` (`PATCH /lists/{id}`)

- 先に`SELECT user_id FROM lists WHERE id = ?`で所有者チェックし、存在しない/他ユーザーのものなら`404 LIST_NOT_FOUND`
- `name`が送られていれば更新。`tasks.go`の`Update`と違って他に更新可能なフィールドがないため、SET句の動的組み立ては行っていない

### `Delete` (`DELETE /lists/{id}`)

- `DELETE FROM lists WHERE id = ? AND user_id = ?`。`RowsAffected() == 0`なら`404 LIST_NOT_FOUND`
- 所属していた`tasks.list_id`は`schema.sql`の`ON DELETE SET NULL`によりDB側で自動的に`NULL`に戻る(タスク自体は削除されない、[ADR-0021](../../../adr/0021-task-list-grouping.md))。アプリケーションコード側では何もしなくてよい

## 読むときのポイント

- `tasks.go`の各ハンドラと同じ設計パターン(所有者チェックを`WHERE user_id = ?`で兼ねる、他ユーザーのリソースは`404`)を踏襲しており、新しいイディオムは登場しない
- `list_id`をタスク側に持たせる・外す処理(`LIST_ID_NOT_ALLOWED_FOR_CHILD_TASK`の検証含む)は`tasks.go`の`Create`/`Update`にある。このファイル自体は`lists`テーブルのCRUDのみを担当する
