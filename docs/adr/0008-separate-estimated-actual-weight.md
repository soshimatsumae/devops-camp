# 0008: 想定の大変さと実績の大変さを別カラムに分離する

## Status

Accepted

## Context

アプリのテーマは「タスクごとに大変さの想定と実績を記録し、見立ての精度を上げる」こと。この差分を後から比較できる形でデータを持つ必要がある。

## Decision

`tasks`テーブルに`estimated_weight`(作成時に入力、必須、1〜5)と`actual_weight`(完了時に入力、NULL許容、1〜5)を別カラムとして持たせる。`status`を`done`に変更する際に`actual_weight`が未指定だと`ACTUAL_WEIGHT_REQUIRED`エラーで拒否する(`internal/handlers/tasks.go`のUpdate)。

## Consequences

- 見積もりと実績の差分をSQLレベルで単純な引き算・比較として扱える
- タスク完了のたびに実績値の入力を要求するUXになるため、フロントエンド実装時(Step6)に完了操作と実績入力を1つのフローとして設計する必要がある
- 値域(1〜5)のバリデーションはAPI層でのみ行っており、DBスキーマ側では`CHECK`制約を設けていない。DB側の整合性を厳密にしたい場合は追加のマイグレーションが必要
