# 0019: 見積もり・実績の単位を5段階評価から時間ベースの工数に変更する

## Status

Accepted

## Context

[ADR-0008](./0008-separate-estimated-actual-weight.md)で導入した`estimated_weight`/`actual_weight`は「大変さ」を1〜5の5段階評価で記録する設計だった。Step5でバックエンド改修を進める中で、「重さ(主観的な5段階評価)ではなく、時間で評価したい」という要望が出た。見積もりと実績の差分を振り返るという本アプリのテーマ([README.md](../../README.md))に照らすと、「1→3に見積もりがズレた」という評価より「2時間の見積もりが5時間かかった」という評価の方が、差分の大きさを具体的に把握しやすい。

## Decision

`tasks`テーブルの`estimated_weight`/`actual_weight`(`SMALLINT`、1〜5)を、`estimated_hours`/`actual_hours`(`DECIMAL(6,1)`、時間単位の小数)に列名・型とも変更する。

- 単位: 時間。0.5時間刻みの小数を許容する(例: 1.5)
- 値域: 最小0.5時間〜最大9999時間。API層(`internal/handlers/tasks.go`の`validHours`)でこの範囲外の値を`422 VALIDATION_ERROR`で拒否する
- エラーコードも`ESTIMATED_WEIGHT_REQUIRED`/`ACTUAL_WEIGHT_REQUIRED`から`ESTIMATED_HOURS_REQUIRED`/`ACTUAL_HOURS_REQUIRED`に改名する
- `GET /tasks/calendar`(TASK-06)の集計値も`SUM(actual_weight)`から`SUM(actual_hours)`に変更し、レスポンスの`total_weight`は`total_hours`に改名する

列名ごと改名したのは、「重さ」という列名のまま値だけ時間に変えると列名と実態がずれて誤解を招くため。アプリが未リリースでAPI利用者もいないため、破壊的変更のコストは低いと判断した。

## Consequences

- `estimated_weight`/`actual_weight`を参照していたAPI仕様書・DB仕様書・ER図・画面一覧・ADR-0008・既存のGo実装(ハンドラ、モデル、テスト)をすべて更新する必要がある
- 5段階評価という上限のある指標から、実質上限のない時間指標に変わるため、見積もり精度の振り返り(差分表示UIなど)も「何段階ズレたか」ではなく「何時間ズレたか」という表現に変わる
- ヒートマップの集計値(1日あたりの合計)も「重さの合計」から「作業時間の合計」に意味が変わり、色の濃淡の基準(将来的に閾値を決める場合)も時間ベースで設計し直す必要がある
- スキーマ変更(`SMALLINT`→`DECIMAL(6,1)`、列名変更)を伴うため、既存のローカルDBを使っている場合は`schema.sql`を再適用(テーブル再作成)するか、手動で`ALTER TABLE`する必要がある
