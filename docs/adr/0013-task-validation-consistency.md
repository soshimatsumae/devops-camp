# 0013: タスク作成と更新のバリデーション・completed_atの挙動を一致させる

## Status

Accepted

## Context

Step3提出後のレビューで、`internal/handlers/tasks.go`の`Create`と`Update`の間にバリデーションの非対称性があると指摘を受けた。

- `Create`は`title`の空文字を拒否するが、`Update`では`title`を送っても空文字チェックがなく通ってしまう
- `due_date`はAPI仕様書上`YYYY-MM-DD`形式のはずだが、`Create`・`Update`とも形式検証をせず生の文字列をDBへ渡していた
- タスクを`done`にすると`completed_at`が自動セットされる([ADR-0008](./0008-separate-estimated-actual-weight.md))が、`done`から`todo`/`in_progress`に戻したときに`completed_at`をクリアする処理がなく、「未完了なのに完了日時が残る」という矛盾が起きうる

## Decision

- `Update`にも`Create`と同じ「トリム後空文字なら`422 VALIDATION_ERROR`」のチェックを追加する
- `due_date`を`time.Parse("2006-01-02", ...)`で検証し、不正な形式なら`Create`・`Update`とも`422 VALIDATION_ERROR`を返す
- `Update`で`status`が`done`から`done`以外に変わる瞬間(`uncompletingNow`)だけ、`completed_at`を`NULL`に戻す。カレンダー集計([ADR-0010](./0010-calendar-date-range-params.md))は`status = 'done'`のみを見るため集計結果への影響はないが、`GET /tasks/{id}`のレスポンス上の矛盾をなくすために「未完了に戻したらクリアする」を仕様として採用した

## Consequences

- 作成時に必須・禁止されている入力は、更新時にも同じ規則で弾かれるようになり、API利用側から見た挙動の一貫性が上がる
- `docs/API_SPEC.md`のPOST/PATCH双方のエラーレスポンス表に`VALIDATION_ERROR`(due_date形式・title空文字)の行を追加した
- 「未完了に戻したらcompleted_atを保持する」という別の仕様を将来採用する場合は、このADRを`Superseded by`で更新し、新しいADRを追加する
