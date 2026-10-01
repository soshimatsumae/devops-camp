# 0020: タスク一覧の並び替えは`sort`パラメータのホワイトリストで実装する

## Status

Accepted

## Context

Step4の画面設計時に、「並び替えの変更」は`GET /tasks`(TASK-01)の`ORDER BY created_at DESC`が固定のため対応する画面を作れないものとして見送っていた([docs/assignment/step4.md](../assignment/step4.md))。Step5でこれを実装するにあたり、対応する並び替えキーと、ユーザー入力をどうSQLの`ORDER BY`句に反映するかを検討した。

`ORDER BY`句はプレースホルダ(`?`)では値を渡せない(カラム名・方向はSQL構文の一部であり、値として束縛できない)ため、ユーザー入力の文字列をそのまま`ORDER BY`に連結するとSQLインジェクションの通り道になる。

## Decision

`GET /tasks`に`sort`クエリパラメータを追加する。取りうる値を`due_date_asc`/`due_date_desc`/`estimated_hours_asc`/`estimated_hours_desc`の4つに限定し、`internal/handlers/tasks.go`の`validSorts`(`map[string]string`)でこれらの値だけを信頼できる`ORDER BY`句にマッピングする。ホワイトリストにない値は`400 INVALID_SORT`で拒否する。未指定時は従来通り`created_at DESC`。

並び替えキーは「締切日順」(`due_date`)と「見積もり工数順」(`estimated_hours`)の2つに限定し、「実績工数順」(`actual_hours`)は対象外とした(未完了タスクは`actual_hours`が`NULL`のため、並び替えの実用性が低いと判断)。

## Consequences

- ユーザー入力の文字列を直接SQLに連結しないため、`ORDER BY`経由のSQLインジェクションの心配がない
- 対応する並び替えキーを増やす場合は`validSorts`にエントリを追加するだけで済む
- `due_date`は`NULL`許容カラムのため、`due_date_asc`指定時は締切日未設定のタスクがMySQLのデフォルト挙動(`NULL`を最小値として扱う)により先頭に来る。これが直感に反する場合は、将来的に`due_date IS NULL`を使った並び替えの調整を検討する
